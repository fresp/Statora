package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var database *mongo.Database

func ConnectMongo(uri, dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	c, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}

	if err = c.Ping(ctx, nil); err != nil {
		return err
	}

	client = c
	database = c.Database(dbName)
	log.Printf("Connected to MongoDB: %s/%s", uri, dbName)
	return nil
}

func GetDB() *mongo.Database {
	return database
}

func GetCollection(name string) *mongo.Collection {
	return database.Collection(name)
}

func DisconnectMongo() {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}
}
