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

	// If the environment is PRODUCTION, do not load the .env file
	if environment == "PRODUCTION" {
		return
	}

	// Load the .env file if the environment is not production
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Globals.Environment = environment
	Globals.Loaded = true
}

// GetRunningEnvironment returns the value of the ENVIRONMENT environment variable
func GetRunningEnvironment() string {
	if Globals.Loaded == false {
		load()
	}

	return Globals.Environment
}

// GetEnvironmentVariable returns the value of the environment variable with the given name
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

// GetAccessTokenSecret returns the value of the ACCESS_TOKEN_SECRET environment variable and update the global variable if it is empty
func GetAccessTokenSecret() string {
	if Globals.AccessTokenSecret == "" {
		accessTokenSecret := GetEnvironmentVariable("ACCESS_TOKEN_SECRET")
		Globals.AccessTokenSecret = accessTokenSecret
	}

	return Globals.AccessTokenSecret
}

// GetWsTokenSecret returns the value of the WS_TOKEN_SECRET environment variable and update the global variable if it is empty
func GetWsTokenSecret() string {
	if Globals.WsTokenSecret == "" {
		wsTokenSecret := GetEnvironmentVariable("WS_TOKEN_SECRET")
		Globals.WsTokenSecret = wsTokenSecret
	}

	return Globals.WsTokenSecret
}

// GetRefreshTokenSecret returns the value of the REFRESH_TOKEN_SECRET environment variable and update the global variable if it is empty
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

// GetLoomiesGenerationTimeouts returns the values of the GAME_MIN_LOOMIES_GENERATION_TIMEOUT and GAME_MAX_LOOMIES_GENERATION_TIMEOUT environment variables and update the global variables if they are empty
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

// GetLoomiesGenerationAmounts returns the values of the GAME_MIN_LOOMIES_GENERATION_AMOUNT and GAME_MAX_LOOMIES_GENERATION_AMOUNT environment variables and update the global variables if they are empty
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

// GetLoomiesGenerationRadius returns the value of the GAME_LOOMIES_GENERATION_RADIUS environment variable and update the global variable if it is empty
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

// GetMaxLoomiesPerZone returns the value of the GAME_MAX_LOOMIES_PER_ZONE environment variable and update the global variable if it is empty
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

// GetLoomiesExperienceParameters Returns the values of the GAME_LOOMIE_MIN_REQUIRED_EXPERIENCE and GAME_LOOMIE_EXPERIENCE_FACTOR environment variables and update the global variables if they are empty
func GetLoomiesExperienceParameters() (float64, float64) {
	if Globals.MinLoomieRequiredExperience == 0 || Globals.LoomieExperienceFactor == 0 {
		// Get values (as strings) from the environment
		minLoomieRequiredExperienceString := GetEnvironmentVariable("GAME_LOOMIE_MIN_REQUIRED_EXPERIENCE")
		loomieExperienceFactorString := GetEnvironmentVariable("GAME_LOOMIE_EXPERIENCE_FACTOR")

		// Convert the strings to integers
		minLoomieRequiredExperience, _ := strconv.ParseFloat(minLoomieRequiredExperienceString, 64)
		loomieExperienceFactor, _ := strconv.ParseFloat(loomieExperienceFactorString, 64)

		// Set the values in the globals
		Globals.MinLoomieRequiredExperience = minLoomieRequiredExperience
		Globals.LoomieExperienceFactor = loomieExperienceFactor
	}

	return Globals.MinLoomieRequiredExperience, Globals.LoomieExperienceFactor
}

// getMongoClient returns a MongoDB client
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

// ConnectToMongoCollection Connects to a MongoDB collection and returns it
func ConnectToMongoCollection(collectionName string) *mongo.Collection {
	client := getMongoClient()
	return client.Database(os.Getenv("MONGO_DATABASE")).Collection(collectionName)
}
