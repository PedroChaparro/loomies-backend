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

	Globals.Loaded = true
}

func getEnvironmentVariable(name string) string {
	if Globals.Loaded == false {
		load()
	}

	value := os.Getenv(name)

	if value == "" {
		log.Fatal(name + " not set")
	}

	return value
}

func GetAccessTokenSecret() string {
	if Globals.AccessTokenSecret == "" {
		accessTokenSecret := getEnvironmentVariable("ACCESS_TOKEN_SECRET")
		Globals.AccessTokenSecret = accessTokenSecret
	}

	return Globals.AccessTokenSecret
}

func GetRefreshTokenSecret() string {
	if Globals.RefreshTokenSecret == "" {
		refreshTokenSecret := getEnvironmentVariable("REFRESH_TOKEN_SECRET")
		Globals.RefreshTokenSecret = refreshTokenSecret
	}

	return Globals.RefreshTokenSecret
}

// connectToMongo returns a MongoDB client
func getMongoClient() *mongo.Client {
	// Create the connection if it does not exist
	if Globals.MongoClient == nil {
		load()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Connect to mongo db using the user and password from the .env file
		user := getEnvironmentVariable("MONGO_USER")
		password := getEnvironmentVariable("MONGO_PASSWORD")
		hosts := getEnvironmentVariable("MONGO_HOSTS")
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
