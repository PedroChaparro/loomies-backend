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

// TestGetUserItemsSuccess Test the `/user/items` endpoint
func TestGetUserItemsSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Setup the router
	router := tests.SetupGinRouter()
	router.GET("/user/items", middlewares.MustProvideAccessToken(), HandleGetItems)
	router.POST("/gyms/claim-rewards", middlewares.MustProvideAccessToken(), HandleClaimReward)

	// -------------------------
	// 1. Test with no items
	// -------------------------
	w, req := tests.SetupGetRequest("/user/items", tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check response fields
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("User items were successfully retreived", response["message"])
	c.NotNil(response["items"])
	// By default, the user has no items
	c.Equal(0, len(response["items"].([]interface{})))
	c.NotNil(response["loomballs"])
	// By default, the user has no loomballs
	c.Equal(0, len(response["loomballs"].([]interface{})))

	// -------------------------
	// 2. Test with items
	// -------------------------

	// Get an existing gym to claim it's rewards
	var gym interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gym)
	c.NoError(err)

	// Claim the gym's rewards
	w, req = tests.SetupPayloadedRequest("/gyms/claim-rewards", "POST", map[string]interface{}{
		"gym_id":    gym.Id.Hex(),
		"latitude":  gym.Latitude,
		"longitude": gym.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	claimedRewards := response["reward"].([]interface{})

	// Get the user's items
	w, req = tests.SetupGetRequest("/user/items", tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check response fields
	currentUserItems := response["items"].([]interface{})
	currentUserLoomballs := response["loomballs"].([]interface{})
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("User items were successfully retreived", response["message"])
	c.NotNil(response["items"])
	c.NotNil(response["loomballs"])
	c.Equal(len(claimedRewards), (len(currentUserItems) + len(currentUserLoomballs)))

	// Check the quantities of the items to match with the claimed rewards
	for _, claimedReward := range claimedRewards {
		claimedRewardObject := claimedReward.(map[string]interface{})
		found := false
		quantitiesMatch := false

		// Check if the claimed reward is in the user's items
		for _, userItem := range currentUserItems {
			userItemObject := userItem.(map[string]interface{})
			if userItemObject["_id"] == claimedRewardObject["id"] {
				found = true
				quantitiesMatch = userItemObject["quantity"] == claimedRewardObject["quantity"]
				break
			}
		}

		// Check if the claimed reward is in the user's loomballs
		if !found {
			for _, userLoomball := range currentUserLoomballs {
				userLoomballObject := userLoomball.(map[string]interface{})
				if userLoomballObject["_id"] == claimedRewardObject["id"] {
					found = true
					quantitiesMatch = userLoomballObject["quantity"] == claimedRewardObject["quantity"]
					break
				}
			}
		}

		c.Truef(found, "Item with id %s was not found in the user's items or loomballs", claimedRewardObject["id"])
		c.Truef(quantitiesMatch, "Item with id %s has a different quantity than the claimed reward", claimedRewardObject["id"])
	}

	// Check items fields
	for _, item := range response["items"].([]interface{}) {
		itemObject := item.(map[string]interface{})
		c.NotEmpty(itemObject["_id"])
		c.NotEmpty(itemObject["name"])
		c.NotEmpty(itemObject["serial"])
		c.Positive(itemObject["serial"])
		c.NotEmpty(itemObject["description"])
		c.NotEmpty(itemObject["target"])
		c.Contains([]string{"Loomie"}, itemObject["target"])
		c.NotEmpty(itemObject["is_combat_item"])
		c.NotEmpty(itemObject["quantity"])
		c.Positive(itemObject["quantity"])
	}

	// Check loomballs fields
	for _, loomball := range response["loomballs"].([]interface{}) {
		loomballObject := loomball.(map[string]interface{})
		c.NotEmpty(loomballObject["_id"])
		c.NotEmpty(loomballObject["name"])
		c.NotEmpty(loomballObject["serial"])
		c.Positive(loomballObject["serial"])
		c.NotEmpty(loomballObject["quantity"])
		c.Positive(loomballObject["quantity"])
	}

	// Delete the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.Nil(err)
}

// TestGetUserItemsBadRequest Test the `/user/items` endpoint with a bad request
func TestGetUserItemsBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Delete the user
	err := tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.Nil(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.GET("/user/items", middlewares.MustProvideAccessToken(), HandleGetItems)

	// Make the request
	w, req := tests.SetupGetRequest("/user/items", tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check response fields
	c.Equal(404, w.Code)
	c.Equal(true, response["error"])
	c.Equal("User was not found", response["message"])
}
