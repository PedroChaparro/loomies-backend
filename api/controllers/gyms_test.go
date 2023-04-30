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
	c.False(responseGym["was_reward_claimed"].(bool))
	if responseGym["owner"] != nil {
		c.NotEmpty(responseGym["owner"])
	} else {
		c.Nil(responseGym["owner"])
	}

	// By default, the gym should have 6 protectors
	gymProtectors := responseGym["protectors"].([]interface{})
	c.Equal(len(gym.Protectors), len(gymProtectors))
	c.Greater(len(gymProtectors), 0)

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

	// Check if the rewards were added to the user
	var user interfaces.User
	err = models.UserCollection.FindOne(ctx, bson.M{
		"_id": randomUser.Id,
	}).Decode(&user)
	c.NoError(err)

	// Check the rewards quantities
	c.Equal(len(rewards), len(user.Items))

	// For each gym reward
	for _, gymReward := range rewards {
		// Match the user reward
		for _, userReward := range user.Items {
			gymRewardObject := gymReward.(map[string]interface{})

			// Check the quantities
			if gymRewardObject["id"] == userReward.ItemId.Hex() {
				c.Equal(int(gymRewardObject["quantity"].(float64)), userReward.ItemQuantity)
			}

			break
		}
	}

	// Remove user items
	_, err = models.UserCollection.UpdateByID(ctx, user.Id, bson.M{
		"$set": bson.M{
			"items": []interfaces.InventoryItem{},
		},
	})
	c.NoError(err)

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

	// Check if the rewards were added to the user
	err = models.UserCollection.FindOne(ctx, bson.M{
		"_id": randomUser.Id,
	}).Decode(&user)
	c.NoError(err)

	c.Equal(len(rewards), len(user.Items))

	for _, gymReward := range rewards {
		for _, userReward := range user.Items {
			gymRewardObject := gymReward.(map[string]interface{})

			if gymRewardObject["id"] == userReward.ItemId.Hex() {
				c.Equal(int(gymRewardObject["quantity"].(float64)), userReward.ItemQuantity)
			}

			break
		}
	}

	// Remove user items
	_, err = models.UserCollection.UpdateByID(ctx, user.Id, bson.M{
		"$set": bson.M{
			"items": []interfaces.InventoryItem{},
		},
	})
	c.NoError(err)

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

// TestUpdateProtectorsErrors Tests the `/gyms/update-protectors“ endpoint with a bad request
func TestUpdateProtectorsErrors(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Get an existing gym from the end of the database collection
	var gym interfaces.Gym
	err := models.GymsCollection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"$natural": -1})).Decode(&gym)
	c.NoError(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.PUT("/gyms/update-protectors", middlewares.MustProvideAccessToken(), HandleUpdateProtectors)

	// -------------------------
	// 1. Check with an empty payload
	// -------------------------
	w, req := tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", nil, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("JSON payload is invalid or missing", response["message"])

	// -------------------------
	// 2. Check with empty protectors aray
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id":     gym.Id.Hex(),
		"protectors": []string{},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You must add at least one protector", response["message"])

	// -------------------------
	// 3. Check with more than 6 protectors
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id": gym.Id.Hex(),
		"protectors": []string{
			"protector1",
			"protector2",
			"protector3",
			"protector4",
			"protector5",
			"protector6",
			"protector7",
		},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You can't add more than 6 protectors", response["message"])

	// -------------------------
	// 4. Check with non-valid gym id
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id":     "non-valid-id",
		"protectors": []string{"protector1"},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("The gym id is not valid", response["message"])

	// -------------------------
	// 5. Check with non-existing gym id
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id":     primitive.NewObjectID().Hex(),
		"protectors": []string{"protector1"},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(404, w.Code)
	c.Equal(true, response["error"])
	c.Equal("The gym was not found", response["message"])

	// -------------------------
	// 6. Check with a gym that is not owned by the user
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id":     gym.Id.Hex(),
		"protectors": []string{"protector1"},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(403, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You don't own this gym", response["message"])

	// -------------------------
	// 7. Check with invalid protectors ids
	// -------------------------
	// Update the gym owner id
	_, err = models.GymsCollection.UpdateOne(ctx, bson.M{"_id": gym.Id}, bson.M{"$set": bson.M{"owner": randomUser.Id}})
	c.NoError(err)

	w, req = tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id":     gym.Id.Hex(),
		"protectors": []string{"non-valid-id", "non-valid-id2"},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Some of the loomie ids were not valid", response["message"])

	// -------------------------
	// 8. Check with loomies that are not owned by the user
	// -------------------------
	var loomies []interfaces.CaughtLoomie
	cursor, err := models.CaughtLoomiesCollection.Find(ctx, bson.M{}, options.Find().SetLimit(6))
	c.NoError(err)
	err = cursor.All(ctx, &loomies)
	c.NoError(err)
	c.Equal(6, len(loomies))

	w, req = tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id":     gym.Id.Hex(),
		"protectors": []string{loomies[0].Id.Hex(), loomies[1].Id.Hex(), loomies[2].Id.Hex(), loomies[3].Id.Hex(), loomies[4].Id.Hex(), loomies[5].Id.Hex()},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(403, w.Code)
	c.Equal(true, response["error"])
	c.Equal("You don't own all the loomies", response["message"])

	// -------------------------
	// 9. Check with busy loomies
	// -------------------------
	// Update the owner of the loomies
	_, err = models.CaughtLoomiesCollection.UpdateMany(ctx, bson.M{"_id": bson.M{
		"$in": []primitive.ObjectID{loomies[0].Id, loomies[1].Id, loomies[2].Id, loomies[3].Id, loomies[4].Id, loomies[5].Id},
	}}, bson.M{"$set": bson.M{"owner": randomUser.Id}})
	c.NoError(err)

	// Update the busy status of one of the loomies
	_, err = models.CaughtLoomiesCollection.UpdateOne(ctx, bson.M{"_id": loomies[0].Id}, bson.M{"$set": bson.M{"is_busy": true}})
	c.NoError(err)

	w, req = tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id":     gym.Id.Hex(),
		"protectors": []string{loomies[0].Id.Hex(), loomies[1].Id.Hex(), loomies[2].Id.Hex(), loomies[3].Id.Hex(), loomies[4].Id.Hex(), loomies[5].Id.Hex()},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("All the loomies must be free to protect the gym", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}

// TestUpdateProtectorsSuccess Tests the `/gyms/update-protectors“ endpoint with valid data
func TestUpdateProtectorsSuccess(t *testing.T) {
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

	// Get 6 caught loomies
	var loomies []interfaces.CaughtLoomie
	cursor, err := models.CaughtLoomiesCollection.Find(ctx, bson.M{}, options.Find().SetLimit(6))
	c.NoError(err)
	err = cursor.All(ctx, &loomies)
	c.NoError(err)
	c.Equal(6, len(loomies))

	// Update the owner and busy state of the loomies
	_, err = models.CaughtLoomiesCollection.UpdateMany(ctx,
		bson.M{"_id": bson.M{
			"$in": []primitive.ObjectID{loomies[0].Id, loomies[1].Id, loomies[2].Id, loomies[3].Id, loomies[4].Id, loomies[5].Id},
		}},
		bson.M{"$set": bson.M{"owner": randomUser.Id, "is_busy": false}})
	c.NoError(err)

	// Update the gym owner
	_, err = models.GymsCollection.UpdateOne(ctx, bson.M{"_id": gym.Id}, bson.M{"$set": bson.M{"owner": randomUser.Id}})
	c.NoError(err)

	// Update the busy state of one of the loomies and set it as a protector
	_, err = models.CaughtLoomiesCollection.UpdateOne(ctx, bson.M{"_id": loomies[0].Id}, bson.M{"$set": bson.M{"is_busy": true}})
	c.NoError(err)

	_, err = models.GymsCollection.UpdateOne(ctx, bson.M{"_id": gym.Id}, bson.M{"$set": bson.M{"protectors": []primitive.ObjectID{loomies[0].Id}}})
	c.NoError(err)

	// Setup the router
	router := tests.SetupGinRouter()
	router.PUT("/gyms/update-protectors", middlewares.MustProvideAccessToken(), HandleUpdateProtectors)

	// -------------------------
	// Send the request
	// -------------------------
	w, req := tests.SetupPayloadedRequest("/gyms/update-protectors", "PUT", map[string]interface{}{
		"gym_id":     gym.Id.Hex(),
		"protectors": []string{loomies[0].Id.Hex(), loomies[1].Id.Hex(), loomies[2].Id.Hex(), loomies[3].Id.Hex(), loomies[4].Id.Hex(), loomies[5].Id.Hex()},
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Gym protectors were successfully updated", response["message"])

	// Check the gym protectors
	err = models.GymsCollection.FindOne(ctx, bson.M{"_id": gym.Id}).Decode(&gym)
	c.NoError(err)
	c.Equal(6, len(gym.Protectors))

	for index := range gym.Protectors {
		c.Equal(loomies[index].Id, gym.Protectors[index])
	}

	// Check the loomies busy state
	cursor, err = models.CaughtLoomiesCollection.Find(ctx, bson.M{"_id": bson.M{"$in": gym.Protectors}})
	c.NoError(err)
	err = cursor.All(ctx, &loomies)
	c.NoError(err)
	c.Equal(6, len(loomies))

	for index := range loomies {
		c.Equal(true, loomies[index].IsBusy)
	}

	// Remove the user
	err = tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}
