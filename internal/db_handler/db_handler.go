package db_handler

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	m "CinemaCityBrowser/internal/model"
)

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

func (mh *MongoHandler) upsertMany(collectionName string, arr []m.WithIndex) mongo.UpdateResult {
	res := mongo.UpdateResult{}
	opts := options.Update().SetUpsert(true)
	coll := mh.db.Collection(collectionName)
	for _, v := range arr {
		update := bson.D{{"$set", v}}
		result, err := coll.UpdateByID(mh.ctx, v.GetId(), update, opts)
		if err != nil {
			panic(err)
		}
		res.MatchedCount += result.MatchedCount
		res.ModifiedCount += result.ModifiedCount
		res.UpsertedCount += result.UpsertedCount
	}
	return res
}

func (mh *MongoHandler) UpsertCinemas(cinemas []m.Cinema) mongo.UpdateResult {
	xs := make([]m.WithIndex, 0)
	for _, cinema := range cinemas {
		xs = append(xs, cinema)
	}
	return mh.upsertMany("cinemas", xs)
}

func (mh *MongoHandler) UpsertFilms(films []m.Film) mongo.UpdateResult {
	xs := make([]m.WithIndex, 0)
	for _, film := range films {
		xs = append(xs, film)
	}
	return mh.upsertMany("film", xs)
}

func (mh *MongoHandler) UpsertEvents(events []m.Event) mongo.UpdateResult {
	xs := make([]m.WithIndex, 0)
	for _, event := range events {
		xs = append(xs, event)
	}
	return mh.upsertMany("events", xs)
}

func (mh *MongoHandler) Close() {
	mh.client.Disconnect(mh.ctx)
}
