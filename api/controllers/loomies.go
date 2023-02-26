package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/mroth/weightedrand/v2"
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
	baseLoomies, err := models.GetBaseLoomies()   // All the possible loomies to generate
	generatedLoomies := []interfaces.WildLoomie{} // The loomies that will be generated

	if err != nil {
		fmt.Println("Error...", err)

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Error getting the base loomies. Please try again later.",
		})
		return
	}

	// Get the amount of loomies to generate between the min and max
	minAmount, maxAmount := configuration.GetLoomiesGenerationAmounts()
	loomiesAmount := utils.GetRandomInt(minAmount, maxAmount)

	weightedChooses := []weightedrand.Choice[interfaces.BaseLoomiesWithPopulatedRarity, int]{}
	for _, loomie := range baseLoomies {
		// The chance is a float between 0 and 1, so we multiply it by 100 to get a percentage
		chance := int(loomie.PopulatedRarity.SpawnChance * 100)
		weightedChooses = append(weightedChooses, weightedrand.NewChoice(loomie, chance))
	}

	weightedChooser, _ := weightedrand.NewChooser(
		weightedChooses...,
	)

	for i := 0; i < loomiesAmount; i++ {
		result := weightedChooser.Pick()

		// Get random coordinates to spawn the new loomie
		randomCoordinates := utils.GetRandomCoordinatesNear(coordinates)

		fmt.Printf("Picked: %v \n", gin.H{
			"Name":   result.Name,
			"Rarity": result.PopulatedRarity.Name,
		})

		wildLoomie := interfaces.WildLoomie{
			Name:   result.Name,
			Types:  result.Types,
			Rarity: result.Rarity,
			// Randomly increase or decrease the stats
			HP:        result.BaseHp + utils.GetRandomInt(-5, 5),
			Attack:    result.BaseAttack + utils.GetRandomInt(-5, 5),
			Defense:   result.BaseDefense + utils.GetRandomInt(-5, 5),
			Latitude:  randomCoordinates.Latitude,
			Longitude: randomCoordinates.Longitude,
		}

		generatedLoomies = append(generatedLoomies, wildLoomie)
	}

	// 4. Update the generation time and timeout in the user doc
	minTimeout, maxTimeout := configuration.GetLoomiesGenerationTimeouts()
	randomTimeout := utils.GetRandomInt(minTimeout, maxTimeout)
	err = models.UpdateUserGenerationTimes(id.(string), currentTimestamp, 0)

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
		"loomies": generatedLoomies,
		"time":    currentTimestamp,
		"timeout": randomTimeout,
	})
}
