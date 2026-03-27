package database

import (
	"context"
	"fmt"
	"log"
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

	// 🔥 Extract DB name (tanpa url.Parse)
	dbName := extractDBName(uri)
	if dbName == "" {
		return fmt.Errorf("database name cannot be empty (missing /dbname in URI)")
	}

	database = c.Database(dbName)

	log.Printf("Connected to MongoDB (db=%s)", dbName)

	return nil
}

// 🔥 safe untuk replica set URI
func extractDBName(uri string) string {
	// ambil bagian setelah host
	parts := strings.Split(uri, "/")
	if len(parts) < 4 {
		return ""
	}

	// bagian setelah "/" terakhir sebelum query
	dbPart := parts[3]

	// buang query param
	if i := strings.Index(dbPart, "?"); i != -1 {
		dbPart = dbPart[:i]
	}

	return dbPart
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