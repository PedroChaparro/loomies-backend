package configuration

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Globals = interfaces.Globals{}
var client *mongo.Client

// load loads the .env file if the environment is not production and create connections
func load() {
	environment := os.Getenv("ENVIRONMENT")

	// If the environment is production, do not load the .env file
	if environment == "production" {
		return
	}

	// Load the .env file if the environment is not production
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// connectToMongo returns a MongoDB client
func getMongoClient() *mongo.Client {
	// Create the connection if it does not exist
	if Globals.MongoClient == nil {
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

		Globals.MongoClient = client
	}

	return Globals.MongoClient
}

// connectToMongoCollection returns a MongoDB collection
func ConnectToMongoCollection(collectionName string) *mongo.Collection {
	client = getMongoClient()
	return client.Database(os.Getenv("MONGO_DATABASE")).Collection(collectionName)
}
