package configuration

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Load loads the .env file if the environment is not production and create connections
func Load() {
	environment := os.Getenv("ENVIRONMENT")

	// Load .env file if environment is not production
	if environment != "production" {
		return
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal("Error connecting to MongoDB", err)
	}
}

// ConnectToMongoCollection returns a MongoDB collection
func ConnectToMongoCollection(collectionName string) *mongo.Collection {
	return client.Database(os.Getenv("MONGODB_DATABASE")).Collection(collectionName)
}
