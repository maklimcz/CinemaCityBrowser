package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	cinemasURL        string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/cinemas/with-event/until/3000-06-06?attr=&lang=pl_PL`
	datesURL          string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/dates/in-cinema/1097/until/3000-06-06?attr=&lang=pl_PL`
	eventsURLtemplate string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/film-events/in-cinema/%v/at-date/%v?attr=&lang=pl_PL`
)

type WithIndex interface {
	getId() string
}

type CinemasResponse struct {
	Body struct {
		Cinemas []Cinema
	}
}

type Cinema struct {
	Id   string `bson:"_id"`
	Name string `json:"displayName" bson:"name"`
}

func (c Cinema) getId() string {
	return c.Id
}

type DatesResponse struct {
	Body struct {
		Dates []string
	}
}

type EventsResponse struct {
	Body struct {
		Films  []Film
		Events []Event
	}
}

type Event struct {
	Id            string   `bson:"_id"`
	FilmId        string   `bson:"filmId"`
	CinemaId      string   `bson:"cinemaId"`
	BusinessDay   string   `bson:"businessDay"`
	EventDateTime string   `bson:"eventDateTime"`
	Attributes    []string `json:"attributeIds"`
	BookingLink   string   `bson:"bookingLink"`
	Auditorium    string   `json:"AuditoriumTinyName"`
}

func (e Event) getId() string {
	return e.Id
}

type Film struct {
	Id          string `bson:"_id"`
	Name        string
	Length      int
	PosterLink  string
	Link        string
	ReleaseYear string
}

func (f Film) getId() string {
	return f.Id
}

type HttpHandler struct {
	httpClient *http.Client
}

func (hh *HttpHandler) fetch_url(url string) []byte {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := hh.httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func (hh *HttpHandler) fetch_cinemas() []Cinema {
	body := hh.fetch_url(cinemasURL)
	var resp CinemasResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	cinemas := resp.Body.Cinemas
	return cinemas
}

func (hh *HttpHandler) fetch_dates() []string {
	body := hh.fetch_url(datesURL)
	var resp DatesResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	cinemas := resp.Body.Dates
	return cinemas
}

func (hh *HttpHandler) fetch_events(cinema Cinema, date string) ([]Film, []Event) {
	url := fmt.Sprintf(eventsURLtemplate, cinema.Id, date)
	log.Println(url)
	body := hh.fetch_url(url)
	var resp EventsResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	return resp.Body.Films, resp.Body.Events
}

type MongoHandler struct {
	client *mongo.Client
	ctx    context.Context
	db     *mongo.Database
}

func (mh *MongoHandler) Init() {
	var err error
	mh.client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://root:example@localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	mh.ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = mh.client.Connect(mh.ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = mh.client.Ping(mh.ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	mh.db = mh.client.Database("cinema-city")
}

func (mh *MongoHandler) UpsertMany(collectionName string, arr []WithIndex) mongo.UpdateResult {
	res := mongo.UpdateResult{}
	opts := options.Update().SetUpsert(true)
	coll := mh.db.Collection(collectionName)
	for _, v := range arr {
		update := bson.D{{"$set", v}}
		result, err := coll.UpdateByID(mh.ctx, v.getId(), update, opts)
		if err != nil {
			panic(err)
		}
		res.MatchedCount += result.MatchedCount
		res.ModifiedCount += result.ModifiedCount
		res.UpsertedCount += result.UpsertedCount
	}
	return res
}

func (mh *MongoHandler) UpsertCinemas(cinemas []Cinema) mongo.UpdateResult {
	xs := make([]WithIndex, 0)
	for _, cinema := range cinemas {
		xs = append(xs, cinema)
	}
	return mh.UpsertMany("cinemas", xs)
}

func (mh *MongoHandler) UpsertFilms(films []Film) mongo.UpdateResult {
	xs := make([]WithIndex, 0)
	for _, film := range films {
		xs = append(xs, film)
	}
	return mh.UpsertMany("film", xs)
}

func (mh *MongoHandler) UpsertEvents(events []Event) mongo.UpdateResult {
	xs := make([]WithIndex, 0)
	for _, event := range events {
		xs = append(xs, event)
	}
	return mh.UpsertMany("events", xs)
}

func (mh *MongoHandler) Close() {
	mh.client.Disconnect(mh.ctx)
}

func main() {
	hh := HttpHandler{httpClient: &http.Client{Timeout: 5 * time.Second}}

	cinemas := hh.fetch_cinemas()
	dates := hh.fetch_dates()

	var mh MongoHandler
	mh.Init()
	defer mh.Close()

	result := mh.UpsertCinemas(cinemas)
	log.Printf("Cinemas: matched=%v, modified=%v, upserted=%v\n", result.MatchedCount, result.ModifiedCount, result.UpsertedCount)

	for _, date := range dates {
		for _, cinema := range cinemas {
			if cinema.Name == "Wrocław - Wroclavia" {
				log.Printf("---Fetching repertoire for cinema %v on %v\n", cinema.Name, date)
				films, events := hh.fetch_events(cinema, date)
				result = mh.UpsertFilms(films)
				log.Printf("Films: matched=%v, modified=%v, upserted=%v\n", result.MatchedCount, result.ModifiedCount, result.UpsertedCount)
				result = mh.UpsertEvents(events)
				log.Printf("Events: matched=%v, modified=%v, upserted=%v\n", result.MatchedCount, result.ModifiedCount, result.UpsertedCount)
			}
		}
	}

	/*

		pipeline := mongo.Pipeline{
			{{"$match", bson.D{
				{"businessDay", "2022-12-27"},
				{"cinemaId", "1097"},
			}}},
			{{"$lookup", bson.D{
				{"from", "films"},
				{"localField", "filmId"},
				{"foreignField", "_id"},
				{"as", "filmDoc"},
			}}},
			{{"$project", bson.D{
				{"_id", 0},
				{"auditorium", 1},
				{"filmName", "$filmDoc.name"},
				{"filmLength", "$filmDoc.length"},
				{"start", "$eventDateTime"},
				{"attributes", 1},
				{"bookingLink", 1},
				{"filmLink", "$filmDoc.link"},
			}}},
			{{"$unwind", bson.D{
				{"path", "$filmName"},
			}}},
			{{"$unwind", bson.D{
				{"path", "$filmLength"},
			}}},
			{{"$sort", bson.D{
				{"auditorium", 1},
				{"start", 1},
			}}},
		}

		db.Collection("fullEvents").Drop(ctx)

		err = db.CreateView(ctx, "fullEvents", "events", pipeline)
		if err != nil {
			log.Fatal(err)
		}

		cursor, err := db.Collection("fullEvents").Find(ctx, bson.D{})
		if err != nil {
			log.Fatal(err)
		}

		var results []bson.M
		if err = cursor.All(context.TODO(), &results); err != nil {
			log.Fatal(err)
		}
		for _, result := range results {
			fmt.Println(result)
		}
	*/

}
