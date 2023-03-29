package controllers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/mroth/weightedrand/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// generateLoomies "private" function to generate loomies for the user
func generateLoomies(userId string, userCoordinates interfaces.Coordinates) error {
	errors := map[string]error{
		"USER_NOT_FOUND":            errors.New("User was not found"),
		"SERVER_BASE_LOOMIES_ERROR": errors.New("Error getting the base loomies. Please try again later."),
		"SERVER_UPDATE_TIMES_ERROR": errors.New("Error updating the user times. Please try again later."),
	}

	// 1. Get the user doc from the database to validate the generation times
	user, err := models.GetUserById(userId)

	if err != nil {
		return errors["USER_NOT_FOUND"]
	}

	// 2. Check if the user can generate loomies
	currentTimestamp := time.Now().Unix()
	currentTime := time.Unix(currentTimestamp, 0)
	previousGenerationTime := time.Unix(user.LastLoomieGenerationTime, 0)
	nextGenerationTime := previousGenerationTime.Add(time.Minute * time.Duration(user.CurrentLoomiesGenerationTimeout))

	// If the current time is before the next generation time, the user can't generate loomies
	if currentTime.Before(nextGenerationTime) {
		// Just return nil, it isn't necessary to return an error (the controller will return the same loomies)
		return nil
	}

	// 3. Generate loomies
	baseLoomies, err := models.GetBaseLoomies() // All the possible loomies to generate

	if err != nil {
		return errors["SERVER_BASE_LOOMIES_ERROR"]
	}

	// Get the amount of loomies to generate between the min and max
	minAmount, maxAmount := configuration.GetLoomiesGenerationAmounts()
	loomiesAmount := utils.GetRandomInt(minAmount, maxAmount)
	weightedChooses := []weightedrand.Choice[interfaces.BaseLoomiesWithPopulatedRarity, int]{}

	// Create the weighted choices
	// Read: https://pkg.go.dev/github.com/mroth/weightedrand/v2
	for _, loomie := range baseLoomies {
		// The chance is a float between 0 and 1, so we multiply it by 100 to get a percentage
		chance := int(loomie.PopulatedRarity.SpawnChance * 100)
		weightedChooses = append(weightedChooses, weightedrand.NewChoice(loomie, chance))
	}

	weightedChooser, _ := weightedrand.NewChooser(
		weightedChooses...,
	)

	// Generate the loomies
	for i := 0; i < loomiesAmount; i++ {
		result := weightedChooser.Pick()

		// Get random coordinates to spawn the new loomie
		randomCoordinates := utils.GetRandomCoordinatesNear(userCoordinates)

		/* fmt.Printf("Picked: %v \n", gin.H{
			"Name":   result.Name,
			"Rarity": result.PopulatedRarity.Name,
		}) */

		wildLoomie := interfaces.WildLoomie{
			Serial:    result.Serial,
			Name:      result.Name,
			Types:     result.Types,
			Rarity:    result.Rarity,
			Latitude:  randomCoordinates.Latitude,
			Longitude: randomCoordinates.Longitude,
			// Randomly increase or decrease the stats
			HP:      result.BaseHp + utils.GetRandomInt(-5, 5),
			Attack:  result.BaseAttack + utils.GetRandomInt(-5, 5),
			Defense: result.BaseDefense + utils.GetRandomInt(-5, 5),
		}

		// Insert the new loomie in the database
		models.InsertWildLoomie(wildLoomie)
	}

	// 4. Update the generation time and timeout in the user doc
	minTimeout, maxTimeout := configuration.GetLoomiesGenerationTimeouts()
	randomTimeout := utils.GetRandomInt(minTimeout, maxTimeout)
	err = models.UpdateUserGenerationTimes(userId, currentTimestamp, int64(randomTimeout))

	return nil
}

// HandleNearLoomies generates loomies for the user
func HandleNearLoomies(c *gin.Context) {
	// 0. Get the coordinates from the request body
	coordinates := interfaces.Coordinates{}

	if err := c.BindJSON(&coordinates); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Latitude and longitude are required"})
		return
	}

	// 1. Try to generate new loomies
	id, _ := c.Get("userid")
	err := generateLoomies(id.(string), coordinates)

	if err != nil {
		statusCode := http.StatusInternalServerError

		if strings.Split(err.Error(), "_")[0] == "USER" {
			statusCode = http.StatusBadRequest
		}

		c.AbortWithStatusJSON(statusCode, gin.H{
			"error":   true,
			"message": err.Error(),
		})

		return
	}

	// 2. Return the loomies near the user coordinates
	wildLoomies, err := models.GetNearWildLoomies(coordinates)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "No loomies found",
			})
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Error getting the loomies. Please try again later.",
		})

		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Loomies were retrieved successfully",
		"loomies": wildLoomies,
	})
}

func HandleValidateLoomieExists(c *gin.Context) {
	loomie_id := c.Param("id")

	if loomie_id == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "No Loomie ID was provided"})
		return
	}

	_, err := models.ValidateLoomieExists(loomie_id)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Loomie doesn't exists",
		})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"error":     false,
		"message":   "Loomie exists",
		"loomie_id": loomie_id,
	})
}

func HandleCaptureLoomie(c *gin.Context) {
	loomie_req := interfaces.LoomieForm{}
	userid, _ := c.Get("userid")

	if err := c.BindJSON(&loomie_req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Bad request"})
		return
	}

	if loomie_req.LoomieId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "No Loomie ID was provided"})
		return
	}

	zoneRadiusStr := configuration.GetEnvironmentVariable("GAME_ZONE_RADIUS")
	zoneRadius, _ := strconv.ParseFloat(zoneRadiusStr, 64)
	loomie, err := models.ValidateLoomieExists(loomie_req.LoomieId)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": true, "message": "Loomie was not found"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
			return
		}
	}

	if math.Abs(loomie.Latitude-loomie_req.Latitude) > zoneRadius || math.Abs(loomie.Longitude-loomie_req.Longitude) > zoneRadius {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "User is not near the loomie"})
		return
	}

	user, err := models.GetUserById(userid.(string))

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": true, "message": "User was not found"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
			return
		}
	}

	caught_loomies := interfaces.CaughtLoomie{Owner: user.Id,
		IsBusy:  false,
		Serial:  loomie.Serial,
		Name:    loomie.Name,
		Types:   loomie.Types,
		Rarity:  loomie.Rarity,
		HP:      loomie.HP,
		Attack:  loomie.Attack,
		Defense: loomie.Defense}

	caught_loomies_id, err := models.InsertInCaughtLoomies(caught_loomies)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
		return
	}

	err = models.InsertInUserLoomie(user, caught_loomies_id)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Loomie caught",
	})
}
