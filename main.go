package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

/*
const (
	cinemasURL string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/cinemas/with-event/until/2023-12-22?attr=&lang=pl_PL`
	datesURL   string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/dates/in-cinema/1097/until/2023-12-22?attr=&lang=pl_PL`
	eventsURL  string = `https://www.cinema-city.pl/pl/data-api-service/v1/quickbook/10103/film-events/in-cinema/1097/at-date/2022-12-22?attr=&lang=pl_PL`
)
*/

type CinemasResponse struct {
	Body struct {
		Cinemas []Cinema
	}
}

type Cinema struct {
	Id   string `bson:"_id"`
	Name string `json:"displayName" bson:"name"`
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
	Id            string
	FilmId        string
	CinemaId      string
	BusinessDay   string
	EventDateTime string
	Attributes    []string `json:"attributeIds"`
	BookingLink   string
	Auditorium    string `json:"AuditoriumTinyName"`
}

type Film struct {
	Id          string
	Name        string
	Length      int
	PosterLink  string
	ReleaseYear string
}

func fetch_cinemas() []Cinema {
	fpath := `mockdata/cinemas.json`
	log.Println("Opening a file:", fpath)
	content, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatal("Error reading file.", err)
	}
	var resp CinemasResponse
	err = json.Unmarshal(content, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}

	return resp.Body.Cinemas
}

/*
func fetch_dates() []string {
	fpath := `mockdata/dates.json`
	log.Println("Opening a file:", fpath)
	content, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatal("Error reading file.", err)
	}
	var resp DatesResponse
	err = json.Unmarshal(content, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}

	return resp.Body.Dates
}

func fetch_events() ([]Film, []Event) {
	fpath := `mockdata/events.json`
	log.Println("Opening a file:", fpath)
	content, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatal("Error reading file.", err)
	}
	var resp EventsResponse
	err = json.Unmarshal(content, &resp)
	if err != nil {
		log.Fatal("Error parsing content.", err)
	}

	return resp.Body.Films, resp.Body.Events
}
*/

func main() {
	cinemas := fetch_cinemas()
	//dates := fetch_dates()
	//films, event := fetch_events()

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
	//opts := options.Update().SetUpsert(true)

	cinemasCollection := db.Collection("cinemas")

	cins := make([]any, 0)
	for _, v := range cinemas {
		cins = append(cins, v)
	}

	result, err := cinemasCollection.InsertMany(ctx, cins)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Number of documents inserted: %d\n", len(result.InsertedIDs))
}
