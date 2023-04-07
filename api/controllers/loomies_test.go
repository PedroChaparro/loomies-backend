package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/tests"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// TestGetNearLoomiesBadRequest Test the `/loomies/near` endpoint with bad request
func TestGetNearLoomiesBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/loomies/near", middlewares.MustProvideAccessToken(), HandleNearLoomies)

	// Get a valid coordinates from some gym in the database
	var gym interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gym)
	c.NoError(err)

	// -------------------------
	// 1. Test with bad user timeout
	// -------------------------

	// Update the user timeout to 1hr in the database
	_, err = models.UserCollection.UpdateOne(ctx, bson.M{
		"_id": randomUser.Id,
	}, bson.M{
		"$set": bson.M{
			"currentLoomiesGenerationTimeout": 3600,
		},
	})
	c.NoError(err)

	// Use the gym coordinates to get the near loomies
	w, req := tests.SetupPayloadedRequest("/loomies/near", "POST", map[string]interface{}{
		"latitude":  gym.Latitude,
		"longitude": gym.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the fields
	// Note: The response is not an error, but the loomies array is empty
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Loomies were retrieved successfully", response["message"])
	c.Equal(0, len(response["loomies"].([]interface{})))

	// -------------------------
	// 2. Test with nil payload
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/loomies/near", "POST", nil, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the fields
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Latitude and longitude are required", response["message"])

	// -------------------------
	// 2. Test with non existing user
	// -------------------------
	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)

	// Send the request
	w, req = tests.SetupPayloadedRequest("/loomies/near", "POST", map[string]interface{}{
		"latitude":  gym.Latitude,
		"longitude": gym.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the fields
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("User was not found", response["message"])
}

// TestGetNearLoomiesSuccess Test the `/loomies/near` endpoint
func TestGetNearLoomiesSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get a valid coordinates from some gym in the database
	var gym interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gym)
	c.NoError(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/loomies/near", middlewares.MustProvideAccessToken(), HandleNearLoomies)

	// Use the gym coordinates to get the near loomies
	w, req := tests.SetupPayloadedRequest("/loomies/near", "POST", map[string]interface{}{
		"latitude":  gym.Latitude,
		"longitude": gym.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the fields
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Loomies were retrieved successfully", response["message"])
	c.NotEmpty(response["loomies"])
	c.Greater(len(response["loomies"].([]interface{})), 0)

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}

// TestLoomieExistenceValidationSuccess Test the `/loomies/exists/:id` endpoint
func TestLoomieExistenceValidationSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get a valid loomie from the database
	var loomie interfaces.WildLoomie
	err := models.WildLoomiesCollection.FindOne(ctx, bson.M{}).Decode(&loomie)
	c.NoError(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.GET("/loomies/exists/:id", middlewares.MustProvideAccessToken(), HandleValidateLoomieExists)

	// Make the request
	w, req := tests.SetupGetRequest("/loomies/exists/"+loomie.Id.Hex(), tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the fields
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Loomie exists", response["message"])
	c.Equal(loomie.Id.Hex(), response["loomie_id"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}

// TestLoomieExistenceNonSuccess Test the `/loomies/exists/:id` with Not Found and Bad Request responses
func TestLoomieExistenceNonSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Setup the router
	router := tests.SetupGinRouter()
	router.GET("/loomies/exists/:id", middlewares.MustProvideAccessToken(), HandleValidateLoomieExists)

	// -------------------------
	// 1. Test with a non existing ID
	// -------------------------
	w, req := tests.SetupGetRequest("/loomies/exists/5c9f5c9f5c9f5c9f5c9f5c9f", tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the fields
	c.Equal(404, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Loomie doesn't exists", response["message"])

	// Remove the user
	err := tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}
