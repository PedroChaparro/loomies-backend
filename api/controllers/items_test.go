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
		c.Contains([]bool{true, false}, itemObject["is_combat_item"])
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

// TestUseItemErrors Test the `/items/use` endpoint with errors
func TestUseItemErrors(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get items from the database
	var unknownBeverage interfaces.Item
	err := models.ItemsCollection.FindOne(ctx, bson.M{
		"serial": 7,
	}).Decode(&unknownBeverage)
	c.Nil(err)

	var smallAidKit interfaces.Item
	err = models.ItemsCollection.FindOne(ctx, bson.M{
		"serial": 2,
	}).Decode(&smallAidKit)
	c.Nil(err)

	// Get one random loomie from the database
	var loomie interfaces.CaughtLoomie
	err = models.CaughtLoomiesCollection.FindOne(ctx, bson.M{}).Decode(&loomie)
	c.Nil(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/items/use", middlewares.MustProvideAccessToken(), HandleUseItem)

	// -------------------------
	// Test 1: Test with nil payload
	// -------------------------
	w, req := tests.SetupPayloadedRequest("/items/use", "POST", nil, tests.CustomHeader{}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("JSON payload is invalid or missing", response["message"])

	// -------------------------
	// Test 2: Test with empty loomie_id field
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/items/use", "POST", map[string]interface{}{
		"loomie_id": "",
		"item_id":   unknownBeverage.Id,
	}, tests.CustomHeader{}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("A Loomie is required", response["message"])

	// -------------------------
	// Test 3: Test with empty item_id field
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/items/use", "POST", map[string]interface{}{
		"loomie_id": loomie.Id,
		"item_id":   "",
	}, tests.CustomHeader{}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("An item is required", response["message"])

	// -------------------------
	// Test 4: Test with an item that the user doesn't own
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/items/use", "POST", map[string]interface{}{
		"loomie_id": loomie.Id,
		"item_id":   unknownBeverage.Id,
	}, tests.CustomHeader{}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You don't own the given item", response["message"])

	// -------------------------
	// Test 5: Test with a non-supported item
	// -------------------------
	// Add the item to the user inventory
	_, err = models.UserCollection.UpdateOne(ctx, bson.M{
		"_id": randomUser.Id,
	}, bson.M{
		"$push": bson.M{
			"items": bson.M{
				"item_collection": "items",
				"item_id":         smallAidKit.Id,
				"item_quantity":   13,
			},
		},
	})
	c.NoError(err)

	// Make the request
	w, req = tests.SetupPayloadedRequest("/items/use", "POST", map[string]interface{}{
		"loomie_id": loomie.Id,
		"item_id":   smallAidKit.Id,
	}, tests.CustomHeader{}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(404, w.Code)
	c.Equal(true, response["error"])
	c.Equal("The given item was not found or is a combat item", response["message"])

	// -------------------------
	// Test 6: Test with not enough items
	// -------------------------
	// Add the item to the user inventory
	_, err = models.UserCollection.UpdateOne(ctx, bson.M{
		"_id": randomUser.Id,
	}, bson.M{
		"$push": bson.M{
			"items": bson.M{
				"item_collection": "items",
				"item_id":         unknownBeverage.Id,
				"item_quantity":   0,
			},
		}})
	c.NoError(err)

	// Make the request
	w, req = tests.SetupPayloadedRequest("/items/use", "POST", map[string]interface{}{
		"loomie_id": loomie.Id,
		"item_id":   unknownBeverage.Id,
	}, tests.CustomHeader{}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You don't have enough of this item", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.Nil(err)
}

// TestUseItemSuccess tests the success case of `/items/use` endpoint
func TestUseItemSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get the unknown beverage item
	var unknownBeverage interfaces.Item
	err := models.ItemsCollection.FindOne(ctx, bson.M{
		"serial": 7,
	}).Decode(&unknownBeverage)
	c.NoError(err)

	// Get a loomie from the database
	var loomie interfaces.CaughtLoomie
	err = models.CaughtLoomiesCollection.FindOne(ctx, bson.M{}).Decode(&loomie)
	c.NoError(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/items/use", middlewares.MustProvideAccessToken(), HandleUseItem)

	// -------------------------
	// Test 1: Valid request
	// -------------------------
	// Add the item to the user inventory
	_, err = models.UserCollection.UpdateOne(ctx, bson.M{
		"_id": randomUser.Id,
	},
		bson.M{
			"$push": bson.M{
				"items": bson.M{
					"item_collection": "items",
					"item_id":         unknownBeverage.Id,
					"item_quantity":   2,
				},
			}})
	c.NoError(err)

	// Add the loomie to the user loomies
	_, err = models.UserCollection.UpdateOne(ctx, bson.M{
		"_id": randomUser.Id,
	}, bson.M{
		"$push": bson.M{
			"loomies": loomie.Id,
		},
	})
	c.NoError(err)

	// Update the loomie owner
	_, err = models.CaughtLoomiesCollection.UpdateOne(ctx, bson.M{
		"_id": loomie.Id,
	}, bson.M{
		"$set": bson.M{
			"owner": randomUser.Id,
		},
	})
	c.NoError(err)

	// Make the request
	w, req := tests.SetupPayloadedRequest("/items/use", "POST", map[string]interface{}{
		"loomie_id": loomie.Id,
		"item_id":   unknownBeverage.Id,
	}, tests.CustomHeader{}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Item was successfully used", response["message"])

	// Check the user inventory
	var user interfaces.User
	err = models.UserCollection.FindOne(ctx, bson.M{
		"_id": randomUser.Id,
	}).Decode(&user)
	c.NoError(err)
	c.Equal(1, len(user.Items))
	c.Equal(unknownBeverage.Id, user.Items[0].ItemId)
	c.Equal(1, user.Items[0].ItemQuantity)

	// Check the loomie
	var finalLoomie interfaces.CaughtLoomie
	err = models.CaughtLoomiesCollection.FindOne(ctx, bson.M{
		"_id": loomie.Id,
	}).Decode(&finalLoomie)
	c.NoError(err)
	c.Equal(loomie.Level+1, finalLoomie.Level)

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.Nil(err)
}
