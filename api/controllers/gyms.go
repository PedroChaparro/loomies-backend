package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// HandleClaimReward Handles the request to claim a gym reward
func HandleClaimReward(c *gin.Context) {
	// Get user from request context
	userId, _ := c.Get("userid")
	userIdMongo, _ := primitive.ObjectIDFromHex(userId.(string))

	// Parse request body
	payload := interfaces.ClaimGymRewardReq{}

	if err := c.BindJSON(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Invalid request body"})
		return
	}

	if payload.GymID == "" || payload.Latitude == 0 || payload.Longitude == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Gym id, latitude and longitude are required"})
		return
	}

	// 1. Validate the user is near (at most the zone radius) to the gym
	zoneRadiusStr := configuration.GetEnvironmentVariable("GAME_ZONE_RADIUS")
	zoneRadius, _ := strconv.ParseFloat(zoneRadiusStr, 64)
	gym, err := models.GetGymFromID(payload.GymID)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": true, "message": "Gym not found"})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal error when getting gym, please try again later"})
		return
	}

	if math.Abs(gym.Latitude-payload.Latitude) > zoneRadius || math.Abs(gym.Longitude-payload.Longitude) > zoneRadius {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You are too far from the gym"})
		return
	}

	// 2. Validate the user has not claimed the reward yet
	if models.HasUserClaimedReward(gym.RewardsClaimedBy, userIdMongo) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You already claimed the rewards for this gym"})
		return
	}

	// 3. Check if the user is the owner of the gym
	isOwner := gym.Owner == userIdMongo

	var playerRewards []interfaces.GymRewardItem

	if isOwner {
		// fmt.Println("Giving the owner rewards...")
		playerRewards = gym.CurrentOwnerRewards
	} else {
		// fmt.Println("Giving the player rewards...")
		playerRewards = gym.CurrentPlayersRewards
	}

	// 3. Give the reward to the user and add the user to the list of users that have claimed the reward
	err = models.AddItemsToUserInventory(userIdMongo, playerRewards)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal error when adding items to user inventory, please try again later"})
		return
	}

	err = models.RegisterClaimedReward(gym, userIdMongo)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal error when registering claimed reward, please try again later"})
		return
	}

	// 4. Get the details of the items that the user has received
	var rewardsQuantity = make(map[string]int)
	var itemsIds []primitive.ObjectID
	var loomballsIds []primitive.ObjectID

	for _, reward := range playerRewards {
		if reward.RewardCollection == "items" {
			itemsIds = append(itemsIds, reward.RewardId)
		}

		if reward.RewardCollection == "loom_balls" {
			loomballsIds = append(loomballsIds, reward.RewardId)
		}

		rewardsQuantity[reward.RewardId.Hex()] = reward.RewardQuantity
	}

	items, err := models.GetItemsFromIds(itemsIds)

	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal error when getting items details, please try again later"})
		return
	}

	loomballs, err := models.GetLoomballsFromIds(loomballsIds)

	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal error when getting loomballs details, please try again later"})
		return
	}

	// 5. Create an unique array with the items and loomballs\
	var allRewards []gin.H

	for _, item := range items {
		allRewards = append(allRewards, gin.H{
			"id":       item.Id.Hex(),
			"serial":   item.Serial,
			"name":     item.Name,
			"quantity": rewardsQuantity[item.Id.Hex()],
		})
	}

	for _, loomball := range loomballs {
		allRewards = append(allRewards, gin.H{
			"id":       loomball.Id.Hex(),
			"serial":   loomball.Serial,
			"name":     loomball.Name,
			"quantity": rewardsQuantity[loomball.Id.Hex()],
		})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"error": false, "message": "Reward claimed successfully", "reward": allRewards})
}

// HandleGetGyms Handles the request to get a gym details by id
func HandleGetGym(c *gin.Context) {
	// Get the user id from the context
	userId, _ := c.Get("userid")
	userIdMongo, _ := primitive.ObjectIDFromHex(userId.(string))

	// Get the gym id from the request
	gymId := c.Param("id")

	// Parse id into mongodb object id
	gymIdMongo, err := primitive.ObjectIDFromHex(gymId)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Invalid gym id"})
		return
	}

	// Get Gym from database
	gym, err := models.GetPopulatedGymFromId(gymIdMongo, userIdMongo)

	if err != nil {
		if err == mongo.ErrNoDocuments || err.Error() == "EMPTY_RESULTS" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": true, "message": "The gym was not found"})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal error when getting gym, please try again later"})
		return
	}

	// Create the response (This is needed to allow null values on the Owner field)
	response := gin.H{
		"_id":                gym.Id,
		"name":               gym.Name,
		"protectors":         gym.Protectors,
		"was_reward_claimed": gym.WasRewardClaimed,
		"user_owns_it":       gym.UserOwnsIt,
	}

	if gym.Owner == "" {
		response["owner"] = nil
	} else {
		response["owner"] = gym.Owner
	}

	c.IndentedJSON(http.StatusOK, gin.H{"error": false, "message": "Details of the gym were successfully obtained", "gym": response})
}
