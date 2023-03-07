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

// HandleClaimReward handles the request to claim a gym reward
func HandleClaimReward(c *gin.Context) {
	// Get user from request context
	userId, _ := c.Get("userid")
	userId = userId.(string)
	userIdMongo, _ := primitive.ObjectIDFromHex(userId.(string))

	// Parse request body
	payload := interfaces.ClaimGymRewardReq{}

	if err := c.BindJSON(&payload); err != nil {
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
	}

	if math.Abs(gym.Latitude-payload.Latitude) > zoneRadius || math.Abs(gym.Longitude-payload.Longitude) > zoneRadius {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "User is not near the gym"})
		return
	}

	// 2. Validate the user has not claimed the reward yet
	if models.HasUserClaimedReward(gym, userIdMongo) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "User has already claimed the reward"})
		return
	}

	// 3. Give the reward to the user and add the user to the list of users that have claimed the reward
	err = models.AddItemsToUserInventory(userIdMongo, gym.CurrentRewards)

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

	for _, reward := range gym.CurrentRewards {
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
			"name":     item.Name,
			"quantity": rewardsQuantity[item.Id.Hex()],
		})
	}

	for _, loomball := range loomballs {
		allRewards = append(allRewards, gin.H{
			"id":       loomball.Id.Hex(),
			"name":     loomball.Name,
			"quantity": rewardsQuantity[loomball.Id.Hex()],
		})
	}

	c.IndentedJSON(http.StatusOK, gin.H{"error": false, "message": "Reward claimed successfully", "reward": allRewards})
}
