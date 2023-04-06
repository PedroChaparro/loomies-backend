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

// TestClaimGymRewardsSuccess Tests the `/gyms/claim-rewards“ endpoint with a success response
func TestClaimGymRewardsSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get an existing gym
	var gym interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gym)
	c.NoError(err)

	// -------------------------
	// 1. Test with user rewards
	// -------------------------

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/gyms/claim-rewards", middlewares.MustProvideAccessToken(), HandleClaimReward)
	w, req := tests.SetupPayloadedRequest("/gyms/claim-rewards", "POST", map[string]interface{}{
		"gym_id":    gym.Id.Hex(),
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
	c.Equal("Reward claimed successfully", response["message"])
	c.NotEmpty(response["reward"])

	// Check rewards fields
	var userRewardsIds []string
	rewards := response["reward"].([]interface{})

	for _, reward := range rewards {
		reward := reward.(map[string]interface{})
		t.Log(reward)
		c.NotEmpty(reward["id"])
		c.NotEmpty(reward["name"])
		c.Positive(reward["quantity"])
		c.NotEmpty(reward["serial"])
		c.Positive(reward["serial"])
		userRewardsIds = append(userRewardsIds, reward["id"].(string))
	}

	// Check rewards ids to match with the user rewards on the database
	c.Equal(len(rewards), len(userRewardsIds))
	for _, reward := range rewards {
		c.Contains(userRewardsIds, reward.(map[string]interface{})["id"])
	}

	// -------------------------
	// 2. Test with owner rewards
	// -------------------------

	// Update the gym's owner
	_, err = models.GymsCollection.UpdateByID(ctx, gym.Id, bson.M{
		"$set": bson.M{
			"owner": randomUser.Id,
		},
		"$pull": bson.M{
			"rewards_claimed_by": randomUser.Id,
		},
	})
	c.NoError(err)

	// Make the request
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

	// Check the fields
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Reward claimed successfully", response["message"])
	c.NotEmpty(response["reward"])

	// Check rewards fields
	var ownerRewardsIds []string
	rewards = response["reward"].([]interface{})

	for _, reward := range rewards {
		reward := reward.(map[string]interface{})
		t.Log(reward)
		c.NotEmpty(reward["id"])
		c.NotEmpty(reward["name"])
		c.Positive(reward["quantity"])
		c.NotEmpty(reward["serial"])
		c.Positive(reward["serial"])
		ownerRewardsIds = append(ownerRewardsIds, reward["id"].(string))
	}

	// Check rewards ids to match with the owners rewards on the database
	c.Equal(len(rewards), len(ownerRewardsIds))
	for _, reward := range rewards {
		c.Contains(ownerRewardsIds, reward.(map[string]interface{})["id"])
	}

	// Delete the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}

// TestClaimRewardsBadRequest Tests the `/gyms/claim-rewards“ endpoint with a bad request
func TestClaimRewardsBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get an existing gym
	var gym interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}).Decode(&gym)
	c.NoError(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/gyms/claim-rewards", middlewares.MustProvideAccessToken(), HandleClaimReward)

	// -------------------------
	// 1. Check with an empty payload
	// -------------------------
	w, req := tests.SetupPayloadedRequest("/gyms/claim-rewards", "POST", nil, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Invalid request body", response["message"])

	// -------------------------
	// 2. Check with empty fields
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/gyms/claim-rewards", "POST", map[string]interface{}{
		"latitude":  gym.Latitude,
		"longitude": gym.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Gym id, latitude and longitude are required", response["message"])

	w, req = tests.SetupPayloadedRequest("/gyms/claim-rewards", "POST", map[string]interface{}{
		"gym_id": gym.Id.Hex(),
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Gym id, latitude and longitude are required", response["message"])

	// -------------------------
	// 3. Check with non-existing gym id
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/gyms/claim-rewards", "POST", map[string]interface{}{
		"gym_id":    "5f6b9c1b9c9c9c9c9c9c9c9c",
		"latitude":  gym.Latitude,
		"longitude": gym.Longitude,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(404, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Gym not found", response["message"])

	// -------------------------
	// 4. Check with away coordinates
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/gyms/claim-rewards", "POST", map[string]interface{}{
		"gym_id":    gym.Id.Hex(),
		"latitude":  gym.Latitude + 0.0036,
		"longitude": gym.Longitude + 0.0036,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You are too far from the gym", response["message"])

	// -------------------------
	// 5. Check with already claimed rewards
	// -------------------------
	// Insert the user in the gym `rewards_claimed_by` array
	_, err = models.GymsCollection.UpdateOne(ctx, bson.M{
		"_id": gym.Id,
	}, bson.M{
		"$push": bson.M{
			"rewards_claimed_by": randomUser.Id,
		},
	})

	c.NoError(err)

	// Make the request
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

	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You already claimed the rewards for this gym", response["message"])

	// Delete the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}
