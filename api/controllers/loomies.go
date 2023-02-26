package controllers

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
)

// HandleNearLoomies generates loomies for the user
func HandleNearLoomies(c *gin.Context) {
	// 0. Get the coordinates from the request body
	coordinates := interfaces.Coordinates{}

	if err := c.BindJSON(&coordinates); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Latitude and longitude are required"})
		return
	}

	// 1. Get the user doc from the database to validate the generation times
	id, _ := c.Get("userid")
	user, err := models.GetUserById(id.(string))

	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{
			"error":   true,
			"message": "User not found",
		})
		return
	}

	// 2. Check if the user can generate loomies
	currentTimestamp := time.Now().Unix()
	currentTime := time.Unix(currentTimestamp, 0)
	previousGenerationTime := time.Unix(user.LastLoomieGenerationTime, 0)
	nextGenerationTime := previousGenerationTime.Add(time.Minute * time.Duration(user.CurrentLoomiesGenerationTimeout))

	if currentTime.Before(nextGenerationTime) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "You can't generate loomies yet",
		})
		return
	}

	// 3. Generate loomies
	// TODO

	// 4. Update the generation time and timeout in the user doc
	minTimeout, maxTimeout := configuration.GetLoomiesGenerationTimeouts()
	rand.Seed(time.Now().UnixNano())
	randomTimeout := rand.Intn(maxTimeout-minTimeout) + minTimeout
	err = models.UpdateUserGenerationTimes(id.(string), currentTimestamp, int64(randomTimeout))

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Error updating the loomies generation times",
		})
		return
	}

	c.IndentedJSON(200, gin.H{
		"error":   false,
		"message": "Working on it :)",
		"time":    currentTimestamp,
		"timeout": randomTimeout,
	})
}
