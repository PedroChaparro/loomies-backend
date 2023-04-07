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

// TestLoomieCaptureSuccess Test the `/loomies/capture` endpoint
func TestLoomieCaptureSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get the loomballs from the database
	loomballSerials := []int{8, 9, 10}
	var loomballs []interfaces.Loomball

	// Get the loombals and sort them by serial
	cursor, err := models.LoomballsCollection.Find(ctx, bson.M{
		"serial": bson.M{
			"$in": loomballSerials,
		},
	}, options.Find().SetSort(bson.M{
		"serial": 1,
	}))

	c.NoError(err)

	err = cursor.All(ctx, &loomballs)
	c.NoError(err)
	c.Equal(len(loomballs), len(loomballSerials))

	// Set a lot of loomballs for the user
	for _, loomball := range loomballs {
		_, err := models.UserCollection.UpdateOne(ctx, bson.M{
			"_id": randomUser.Id,
		}, bson.M{
			"$push": bson.M{
				"items": bson.M{
					"item_collection": "loom_balls",
					"item_id":         loomball.Id,
					"item_quantity":   3600,
				},
			},
		},
		)

		c.NoError(err)
	}

	// Get the user again from the database
	var user interfaces.User

	err = models.UserCollection.FindOne(ctx, bson.M{
		"_id": randomUser.Id,
	}).Decode(&user)
	c.NoError(err)

	// Check have the expected amount of loomballs
	c.Equal(len(user.Items), len(loomballSerials))

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/loomies/near", middlewares.MustProvideAccessToken(), HandleNearLoomies)
	router.POST("/loomies/capture", middlewares.MustProvideAccessToken(), HandleCaptureLoomie)

	// -------------------------
	// 1. Capture loomie with "Basic" loomball
	// -------------------------
	// Get a valid coordinates from a gym in the database
	var gym interfaces.Gym
	err = models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gym)
	c.NoError(err)

	// Make the requests to generate, at least, 3 wild loomies
	var loomiesStringIds []string
	var loomiesNames []string

	for {
		w, req := tests.SetupPayloadedRequest("/loomies/near", "POST", map[string]interface{}{
			"latitude":  gym.Latitude,
			"longitude": gym.Longitude,
		}, tests.CustomHeader{
			Name:  "Access-Token",
			Value: loginResponse["accessToken"],
		})

		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &response)

		// Check if there are enough loomies
		c.Equal(200, w.Code)
		c.Equal(false, response["error"])

		if len(response["loomies"].([]interface{})) >= 3 {
			// Save the loomies ids
			for _, loomie := range response["loomies"].([]interface{}) {
				loomieObject := loomie.(map[string]interface{})
				loomiesStringIds = append(loomiesStringIds, loomieObject["_id"].(string))
				loomiesNames = append(loomiesNames, loomieObject["name"].(string))
			}

			break
		}
	}

	// Capture the loomie
	for {
		// Retry the request until the loomie is captured
		w, req := tests.SetupPayloadedRequest("/loomies/capture", "POST", map[string]interface{}{
			"loomie_id":   loomiesStringIds[0],
			"loomball_id": user.Items[0].ItemId.Hex(),
			"latitude":    gym.Latitude,
			"longitude":   gym.Longitude,
		}, tests.CustomHeader{
			Name:  "Access-Token",
			Value: loginResponse["accessToken"],
		})

		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &response)

		// Check if the loomie is captured
		c.Equal(200, w.Code)
		c.Equal(false, response["error"])

		if response["was_captured"].(bool) {
			// Check the loomies is in the user loomies array
			err = models.UserCollection.FindOne(ctx, bson.M{
				"_id": randomUser.Id,
			}).Decode(&user)
			c.NoError(err)
			c.Equal(1, len(user.Loomies))

			// Get the loomies details
			var loomie interfaces.UserLoomiesRes
			loomies, err := models.GetLoomiesByIds([]primitive.ObjectID{user.Loomies[0]}, user.Id)
			c.NoError(err)
			loomie = loomies[0]

			// Check the loomies was added to the user array in the database
			c.Equal(len(user.Loomies), 1)
			c.Equal(loomie.Name, loomiesNames[0])
			c.Equal("The loomie was captured", response["message"])

			// Stop trying to capture the loomie
			break
		}
	}

	// -------------------------
	// 2. Capture loomie with "Tier 2" loomball
	// -------------------------
	for {
		// Retry the request until the loomie is captured
		w, req := tests.SetupPayloadedRequest("/loomies/capture", "POST", map[string]interface{}{
			"loomie_id":   loomiesStringIds[1],
			"loomball_id": user.Items[1].ItemId.Hex(),
			"latitude":    gym.Latitude,
			"longitude":   gym.Longitude,
		}, tests.CustomHeader{
			Name:  "Access-Token",
			Value: loginResponse["accessToken"],
		})

		router.ServeHTTP(w, req)
		json.Unmarshal(w.Body.Bytes(), &response)
		c.Equal(200, w.Code)
		c.Equal(false, response["error"])

		if response["was_captured"].(bool) {
			// Check the loomies is in the user loomies array
			err = models.UserCollection.FindOne(ctx, bson.M{
				"_id": randomUser.Id,
			}).Decode(&user)
			c.NoError(err)
			c.Equal(2, len(user.Loomies))

			// Get the loomies details
			var loomie interfaces.UserLoomiesRes
			loomies, err := models.GetLoomiesByIds([]primitive.ObjectID{user.Loomies[1]}, user.Id)
			c.NoError(err)
			loomie = loomies[0]

			// Check the loomies was added to the user array in the database
			c.Equal(len(user.Loomies), 2)
			c.Equal(loomie.Name, loomiesNames[1])
			c.Equal("The loomie was captured", response["message"])

			// Stop trying to capture the loomie
			break
		}
	}

	// -------------------------
	// 3. Capture loomie with "Experimental" loomball
	// -------------------------
	// This loombal has a 100% capture rate, so we don't need to retry the request
	w, req := tests.SetupPayloadedRequest("/loomies/capture", "POST", map[string]interface{}{
		"loomie_id":   loomiesStringIds[2],
		"loomball_id": user.Items[2].ItemId.Hex(),
		"latitude":    gym.Latitude,
		"longitude":   gym.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal(true, response["was_captured"].(bool))

	// Check the loomies is in the user loomies array
	err = models.UserCollection.FindOne(ctx, bson.M{
		"_id": randomUser.Id,
	}).Decode(&user)
	c.NoError(err)
	c.Equal(3, len(user.Loomies))

	// Get the loomies details
	var loomie interfaces.UserLoomiesRes
	loomies, err := models.GetLoomiesByIds([]primitive.ObjectID{user.Loomies[2]}, user.Id)
	c.NoError(err)
	loomie = loomies[0]

	// Check the loomies was added to the user array in the database
	c.Equal(len(user.Loomies), 3)
	c.Equal(loomie.Name, loomiesNames[2])
	c.Equal("The loomie was captured", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}
