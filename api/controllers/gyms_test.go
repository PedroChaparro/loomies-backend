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

// TestGymDetailsSuccess Tests the `/gyms/:id` endpoint
func TestGymDetailsSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get an existing gym id from the database
	var gym interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gym)
	c.NoError(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.GET("/gyms/:id", middlewares.MustProvideAccessToken(), HandleGetGym)

	// Make the request
	w, req := tests.SetupGetRequest("/gyms/"+gym.Id.Hex(), tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// -------------------------
	// 1. Check the basic response
	// -------------------------
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Details of the gym were successfully obtained", response["message"])

	responseGym := response["gym"].(map[string]interface{})
	c.Equal(gym.Id.Hex(), responseGym["_id"])
	c.Equal(gym.Name, responseGym["name"])
	c.Nil(responseGym["owner"])
	c.False(responseGym["was_reward_claimed"].(bool))

	// By default, the gym should have 6 protectors
	gymProtectors := responseGym["protectors"].([]interface{})
	c.Equal(len(gym.Protectors), len(gymProtectors))
	c.Equal(6, len(gym.Protectors))

	// Check the protectors fields
	for _, protector := range gymProtectors {
		protector := protector.(map[string]interface{})
		c.NotEmpty(protector["_id"])
		c.NotEmpty(protector["serial"])
		c.NotEmpty(protector["name"])
		c.NotEmpty(protector["level"])
	}

	// -------------------------
	// 2. Check with the random user as the gym owner
	// -------------------------
	// Update the gym owner
	models.GymsCollection.UpdateByID(ctx, gym.Id, bson.M{
		"$set": bson.M{
			"owner": randomUser.Id,
		},
	})

	// Make the request
	w, req = tests.SetupGetRequest("/gyms/"+gym.Id.Hex(), tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the fields
	responseGym = response["gym"].(map[string]interface{})
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Details of the gym were successfully obtained", response["message"])
	c.Equal(randomUser.Username, responseGym["owner"])

	// -------------------------
	// 2. Check with the user claiming the reward
	// -------------------------
	// Update the reward field
	models.GymsCollection.UpdateByID(ctx, gym.Id, bson.M{
		"$push": bson.M{
			"rewards_claimed_by": randomUser.Id,
		},
	})

	// Make the request
	w, req = tests.SetupGetRequest("/gyms/"+gym.Id.Hex(), tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the fields
	responseGym = response["gym"].(map[string]interface{})
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Details of the gym were successfully obtained", response["message"])
	c.True(responseGym["was_reward_claimed"].(bool))

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}

// TestGymDetailsBadRequest Tests the `/gyms/:id` endpoint with a bad request
func TestGymDetailsBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Setup the router
	router := tests.SetupGinRouter()
	router.GET("/gyms/:id", middlewares.MustProvideAccessToken(), HandleGetGym)

	// -------------------------
	// 1. Check with an invalid gym id
	// -------------------------
	w, req := tests.SetupGetRequest("/gyms/123", tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Invalid gym id", response["message"])

	// -------------------------
	// 2. Check with a non-existing gym id
	// -------------------------
	w, req = tests.SetupGetRequest("/gyms/5f6b9c1b9c9c9c9c9c9c9c9c", tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(404, w.Code)
	c.Equal(true, response["error"])
	c.Equal("The gym was not found", response["message"])

	// Delete the user
	err := tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}
