package configuration

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Globals = TGlobals{}

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

func GetEnvironmentVariable(name string) string {
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
		accessTokenSecret := GetEnvironmentVariable("ACCESS_TOKEN_SECRET")
		Globals.AccessTokenSecret = accessTokenSecret
	}

	return Globals.AccessTokenSecret
}

func GetWsTokenSecret() string {
	if Globals.WsTokenSecret == "" {
		wsTokenSecret := GetEnvironmentVariable("WS_TOKEN_SECRET")
		Globals.WsTokenSecret = wsTokenSecret
	}

	return Globals.WsTokenSecret
}

func GetRefreshTokenSecret() string {
	if Globals.RefreshTokenSecret == "" {
		refreshTokenSecret := GetEnvironmentVariable("REFRESH_TOKEN_SECRET")
		Globals.RefreshTokenSecret = refreshTokenSecret
	}

	return Globals.RefreshTokenSecret
}

// GetWildLoomiesTTL Returns the value of the GAME_WILD_LOOMIES_TTL environment variable and update the global variable if it is empty
func GetWildLoomiesTTL() int {
	if Globals.WildLoomiesTTL == 0 {
		// Get the value (as a string) from the environment
		wildLoomiesTTLString := GetEnvironmentVariable("GAME_WILD_LOOMIES_TTL")

		// Convert the string to an integer
		wildLoomiesTTL, _ := strconv.Atoi(wildLoomiesTTLString)

		// Set the value in the globals
		Globals.WildLoomiesTTL = wildLoomiesTTL
	}

	return Globals.WildLoomiesTTL
}

func GetLoomiesGenerationTimeouts() (int, int) {
	if Globals.MinLoomiesGenerationTimeout == 0 || Globals.MaxLoomiesGenerationTimeout == 0 {
		// Get values (as strings) from the environment
		minLoomiesGenerationTimeoutString := GetEnvironmentVariable("GAME_MIN_LOOMIES_GENERATION_TIMEOUT")
		maxLoomiesGenerationTimeoutString := GetEnvironmentVariable("GAME_MAX_LOOMIES_GENERATION_TIMEOUT")

		// Convert the strings to integers
		minLoomiesGenerationTimeout, _ := strconv.Atoi(minLoomiesGenerationTimeoutString)
		maxLoomiesGenerationTimeout, _ := strconv.Atoi(maxLoomiesGenerationTimeoutString)

		// Set the values in the globals
		Globals.MinLoomiesGenerationTimeout = minLoomiesGenerationTimeout
		Globals.MaxLoomiesGenerationTimeout = maxLoomiesGenerationTimeout
	}

	return Globals.MinLoomiesGenerationTimeout, Globals.MaxLoomiesGenerationTimeout
}

func GetLoomiesGenerationAmounts() (int, int) {
	if Globals.MinLoomiesGenerationAmount == 0 || Globals.MaxLoomiesGenerationAmount == 0 {
		// Get values (as strings) from the environment
		minLoomiesGenerationAmountString := GetEnvironmentVariable("GAME_MIN_LOOMIES_GENERATION_AMOUNT")
		maxLoomiesGenerationAmountString := GetEnvironmentVariable("GAME_MAX_LOOMIES_GENERATION_AMOUNT")

		// Convert the strings to integers
		minLoomiesGenerationAmount, _ := strconv.Atoi(minLoomiesGenerationAmountString)
		maxLoomiesGenerationAmount, _ := strconv.Atoi(maxLoomiesGenerationAmountString)

		// Set the values in the globals
		Globals.MinLoomiesGenerationAmount = minLoomiesGenerationAmount
		Globals.MaxLoomiesGenerationAmount = maxLoomiesGenerationAmount
	}

	return Globals.MinLoomiesGenerationAmount, Globals.MaxLoomiesGenerationAmount
}

func GetLoomiesGenerationRadius() float64 {
	if Globals.LoomiesGenerationRadius == 0 {
		// Get value (as string) from the environment
		loomiesGenerationRadiusString := GetEnvironmentVariable("GAME_LOOMIES_GENERATION_RADIUS")

		// Convert the string to float64
		loomiesGenerationRadius, _ := strconv.ParseFloat(loomiesGenerationRadiusString, 64)

		// Set the value in the globals
		Globals.LoomiesGenerationRadius = loomiesGenerationRadius
	}

	return Globals.LoomiesGenerationRadius
}

func GetMaxLoomiesPerZone() int {
	if Globals.MaxLoomiesPerZone == 0 {
		// Get value (as string) from the environment
		maxLoomiesPerZoneString := GetEnvironmentVariable("GAME_MAX_LOOMIES_PER_ZONE")

		// Convert the string to integer
		maxLoomiesPerZone, _ := strconv.Atoi(maxLoomiesPerZoneString)

		// Set the value in the globals
		Globals.MaxLoomiesPerZone = maxLoomiesPerZone
	}

	return Globals.MaxLoomiesPerZone
}

// connectToMongo returns a MongoDB client
func getMongoClient() *mongo.Client {
	// Create the connection if it does not exist
	if Globals.MongoClient == nil {
		load()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Connect to mongo db using the user and password from the .env file
		user := GetEnvironmentVariable("MONGO_USER")
		password := GetEnvironmentVariable("MONGO_PASSWORD")
		hosts := GetEnvironmentVariable("MONGO_HOSTS")
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
	client := getMongoClient()
	return client.Database(os.Getenv("MONGO_DATABASE")).Collection(collectionName)
}
