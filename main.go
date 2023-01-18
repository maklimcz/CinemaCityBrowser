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

func fetch_url(url string) []byte {
	httpClient := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := httpClient.Do(req)
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

func fetch_cinemas() []Cinema {
	body := fetch_url(cinemasURL)
	var resp CinemasResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	cinemas := resp.Body.Cinemas
	return cinemas
}

func fetch_dates() []string {
	body := fetch_url(datesURL)
	var resp DatesResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	cinemas := resp.Body.Dates
	return cinemas
}

func fetch_events(cinemaId string, date string) ([]Film, []Event) {
	url := fmt.Sprintf(eventsURLtemplate, cinemaId, date)
	log.Println(url)
	body := fetch_url(url)
	var resp EventsResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}
	return resp.Body.Films, resp.Body.Events
}

func UpsertMany[T WithIndex](coll *mongo.Collection, ctx *context.Context, arr []T) mongo.UpdateResult {
	res := mongo.UpdateResult{}
	opts := options.Update().SetUpsert(true)
	for _, v := range arr {
		update := bson.D{{"$set", v}}
		result, err := coll.UpdateByID(*ctx, v.getId(), update, opts)
		if err != nil {
			panic(err)
		}
		res.MatchedCount += result.MatchedCount
		res.ModifiedCount += result.ModifiedCount
		res.UpsertedCount += result.UpsertedCount
	}
	return res
}

func fetch_events_and_upsert(cinema Cinema, date string, db *mongo.Database, ctx *context.Context) {
	log.Printf("Fetching repertoire for cinema %v on date %v\n", cinema.Name, date)
	films, events := fetch_events(cinema.Id, date)

	result := UpsertMany(db.Collection("events"), ctx, events)
	log.Printf("%v\t%v\tEvents:\tmatched=%v\tmodified=%v\tupserted=%v\n", date, cinema.Name, result.MatchedCount, result.ModifiedCount, result.UpsertedCount)
	result = UpsertMany(db.Collection("films"), ctx, films)
	log.Printf("%v\t%v\tFilms:\tmatched=%v\tmodified=%v\tupserted=%v\n", date, cinema.Name, result.MatchedCount, result.ModifiedCount, result.UpsertedCount)
	log.Println("End")
}

func main() {
	cinemas := fetch_cinemas()
	dates := fetch_dates()

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:example@localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("cinema-city")

	result := UpsertMany(db.Collection("cinemas"), &ctx, cinemas)
	log.Printf("Cinemas: matched=%v, modified=%v, upserted=%v\n", result.MatchedCount, result.ModifiedCount, result.UpsertedCount)

	for _, date := range dates {
		for _, cinema := range cinemas {
			if cinema.Name == "Wrocław - Wroclavia" {
				fetch_events_and_upsert(cinema, date, db, &ctx)
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
