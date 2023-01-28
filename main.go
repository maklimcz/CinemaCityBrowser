package main

import (
	"log"
	"time"

	db "CinemaCityBrowser/internal/db_handler"
	api "CinemaCityBrowser/internal/http_handler"
)

func main() {
	var hh api.HttpHandler
	hh.Init(5 * time.Second)

	cinemas := api.FetchCinemas(hh)
	dates := api.FetchDates(hh)

	var mh db.MongoHandler
	mh.Init()
	defer mh.Close()

	result := mh.UpsertCinemas(cinemas)
	log.Printf("Cinemas: matched=%v, modified=%v, upserted=%v\n", result.MatchedCount, result.ModifiedCount, result.UpsertedCount)

	for _, date := range dates {
		for _, cinema := range cinemas {
			if cinema.Name == "Wrocław - Wroclavia" {
				log.Printf("---Fetching repertoire for cinema %v on %v\n", cinema.Name, date)
				films, events := api.FetchEvents(hh, cinema, date)
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
