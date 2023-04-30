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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestRegisterCombatBadRequest Test the error cases for the `/combat/register` endpoint
func TestRegisterCombatBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/combat/register", middlewares.MustProvideAccessToken(), HandleCombatRegister)

	// ---- ---- ---- ----
	// Test 1: Test with an empty payload
	// ---- ---- ---- ----
	w, req := tests.SetupPayloadedRequest("/combat/register", "POST", nil, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("The request body should contain the gym id and the user coordinates", response["message"])

	// ---- ---- ---- ----
	// Test 2: Test with far away coordinates
	// ---- ---- ---- ----
	// Get a gym from the database
	var gymDoc interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gymDoc)
	c.NoError(err)

	// Make the request
	w, req = tests.SetupPayloadedRequest("/combat/register", "POST", map[string]interface{}{
		"gym_id":    gymDoc.Id,
		"latitude":  gymDoc.Latitude + 0.1,
		"longitude": gymDoc.Longitude + 0.1,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You are too far away from the gym", response["message"])

	// ---- ---- ---- ----
	// Test 3: Test with a gym owned by the user
	// ---- ---- ---- ----
	// Update the gym owner
	_, err = models.GymsCollection.UpdateOne(ctx, bson.M{"_id": gymDoc.Id}, bson.M{"$set": bson.M{"owner": randomUser.Id}})
	c.NoError(err)

	// Make the request
	w, req = tests.SetupPayloadedRequest("/combat/register", "POST", map[string]interface{}{
		"gym_id":    gymDoc.Id,
		"latitude":  gymDoc.Latitude,
		"longitude": gymDoc.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You can't challenge your own gym", response["message"])

	// Reset the gym owner
	_, err = models.GymsCollection.UpdateOne(ctx, bson.M{"_id": gymDoc.Id}, bson.M{"$set": bson.M{"owner": nil}})
	c.NoError(err)

	// ---- ---- ---- ----
	// Test 4: Test with no loomies in the loomie team
	// ---- ---- ---- ----
	w, req = tests.SetupPayloadedRequest("/combat/register", "POST", map[string]interface{}{
		"gym_id":    gymDoc.Id,
		"latitude":  gymDoc.Latitude,
		"longitude": gymDoc.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You must have at least one loomie in your team to start a combat.", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}

// TestRegisterCombatSuccess Test the success case for the `/combat/register` endpoint
func TestRegisterCombatSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()
	c.NotEmpty(loginResponse["accessToken"])

	// Get a gym from the database
	var gymDoc interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gymDoc)
	c.NoError(err)

	// Get 6 loomies from the end of the caught loomies collection
	var loomies []interfaces.CaughtLoomie
	cursor, err := models.CaughtLoomiesCollection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"_id": -1}).SetLimit(6))
	c.NoError(err)
	err = cursor.All(ctx, &loomies)
	c.NoError(err)
	c.Equal(6, len(loomies))

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/combat/register", middlewares.MustProvideAccessToken(), HandleCombatRegister)

	// ---- ---- ----
	// Test 1: Test with a valid payload
	// ---- ---- ----
	// Update the loomies owner
	_, err = models.CaughtLoomiesCollection.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": []primitive.ObjectID{loomies[0].Id, loomies[1].Id, loomies[2].Id, loomies[3].Id, loomies[4].Id, loomies[5].Id}}}, bson.M{
		"$set": bson.M{
			"owner":   randomUser.Id,
			"is_busy": false,
		},
	})

	// Update the user loomie team
	_, err = models.UserCollection.UpdateOne(ctx, bson.M{"_id": randomUser.Id}, bson.M{
		"$set": bson.M{
			"loomies":     []primitive.ObjectID{loomies[0].Id, loomies[1].Id, loomies[2].Id, loomies[3].Id, loomies[4].Id, loomies[5].Id},
			"loomie_team": []primitive.ObjectID{loomies[0].Id, loomies[1].Id, loomies[2].Id, loomies[3].Id, loomies[4].Id, loomies[5].Id},
		},
	})
	c.NoError(err)

	// Make the request
	w, req := tests.SetupPayloadedRequest("/combat/register", "POST", map[string]interface{}{
		"gym_id":    gymDoc.Id,
		"latitude":  gymDoc.Latitude,
		"longitude": gymDoc.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Token was created successfully", response["message"])
	c.NotEmpty(response["combat_token"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}
