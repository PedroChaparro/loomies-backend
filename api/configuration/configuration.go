package configuration

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// load loads the .env file if the environment is not production and create connections
func load() {
	environment := os.Getenv("ENVIRONMENT")

	// Load .env file if environment is not production
	if environment == "production" {
		return
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// connectToMongo returns a MongoDB client
func connectToMongo() *mongo.Client {
	load()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to mongo db using the user and password from the .env file
	user := os.Getenv("MONGO_USER")
	password := os.Getenv("MONGO_PASSWORD")
	hosts := os.Getenv("MONGO_HOSTS")
	uri := fmt.Sprintf("mongodb://%s:%s@%s", user, password, hosts)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Error connecting to MongoDB", err)
	}

	return client
}

// connectToMongoCollection returns a MongoDB collection
func ConnectToMongoCollection(collectionName string) *mongo.Collection {
	client = connectToMongo()
	return client.Database(os.Getenv("MONGO_DATABASE")).Collection(collectionName)
}
