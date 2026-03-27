package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var database *mongo.Database

func ConnectMongo(uri string) error {
	if uri == "" {
		return fmt.Errorf("MONGODB_URI is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(5).
		SetServerSelectionTimeout(10 * time.Second).
		SetRetryWrites(true)

	c, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}

	if err = c.Ping(ctx, nil); err != nil {
		return err
	}

	client = c

	// 🔥 Extract DB name dari URI
	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid Mongo URI: %v", err)
	}

	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty (missing /dbname in URI)")
	}

	database = c.Database(dbName)

	log.Printf("Connected to MongoDB (db=%s)", dbName)

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