package controllers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/mroth/weightedrand/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// generateLoomies "private" function to generate loomies for the user
func generateLoomies(userId string, userCoordinates interfaces.Coordinates) error {
	errors := map[string]error{
		"USER_NOT_FOUND":                errors.New("User was not found"),
		"SERVER_BASE_LOOMIES_ERROR":     errors.New("Error getting the base loomies. Please try again later."),
		"SERVER_UPDATE_TIMES_ERROR":     errors.New("Error updating the user times. Please try again later."),
		"SERVER_OUTDATED_LOOMIES_ERROR": errors.New("Error removing the outdated loomies. Please try again later."),
	}

	// Remove the expired loomies before generating new ones
	err := models.RemoveNearExpiredLoomies(userCoordinates)

	if err != nil {
		return errors["SERVER_OUTDATED_LOOMIES_ERROR"]
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
			HP:         result.BaseHp + utils.GetRandomInt(-5, 5),
			Attack:     result.BaseAttack + utils.GetRandomInt(-5, 5),
			Defense:    result.BaseDefense + utils.GetRandomInt(-5, 5),
			Level:      utils.GetRandomLevel(),
			Experience: 0,
			CapturedBy: []primitive.ObjectID{},
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

// HandleNearLoomies Handle the request to get the loomies near the user and generate new loomies as possible
func HandleNearLoomies(c *gin.Context) {
	// 0. Get the coordinates from the request body
	coordinates := interfaces.Coordinates{}

	if err := c.BindJSON(&coordinates); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Latitude and longitude are required"})
		return
	}

	// 1. Try to generate new loomies
	id, _ := c.Get("userid")
	userMongoId, _ := primitive.ObjectIDFromHex(id.(string))
	err := generateLoomies(id.(string), coordinates)

	if err != nil {
		statusCode := http.StatusInternalServerError

		if strings.Split(err.Error(), " ")[0] == "User" {
			statusCode = http.StatusBadRequest
		}

		c.AbortWithStatusJSON(statusCode, gin.H{
			"error":   true,
			"message": err.Error(),
		})

		return
	}

	// 2. Return the loomies near the user coordinates
	wildLoomies, err := models.GetNearWildLoomies(coordinates, userMongoId)

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

// HandleValidateLoomieExists Handle the request to validate if a loomie exists
func HandleValidateLoomieExists(c *gin.Context) {
	loomie_id := c.Param("id")
	_, err := models.GetWildLoomieById(loomie_id)

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

// HandleFuseLoomies Handle the request to fuse two loomies
func HandleFuseLoomies(c *gin.Context) {
	userId, _ := c.Get("userid")

	// Get the loomies ids from the request body
	var req interfaces.FuseLoomiesReq

	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "JSON body is null or invalid"})
		return
	}

	if req.LoomieId1 == "" || req.LoomieId2 == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Both loomies ids are required"})
		return
	}

	// Check the loomies are different
	if req.LoomieId1 == req.LoomieId2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You can't fuse the same loomie"})
		return
	}

	// Check the user owns the loomies
	user, _ := models.GetUserById(userId.(string))
	userMongoId, _ := primitive.ObjectIDFromHex(userId.(string))
	firstLoomieMongoId, _ := primitive.ObjectIDFromHex(req.LoomieId1)
	secondLoomieMongoId, _ := primitive.ObjectIDFromHex(req.LoomieId2)
	var containsLoomie1, containsLoomie2 bool

	for _, id := range user.Loomies {
		if id == firstLoomieMongoId {
			containsLoomie1 = true
		} else if id == secondLoomieMongoId {
			containsLoomie2 = true
		}

		if containsLoomie1 && containsLoomie2 {
			break
		}
	}

	if !containsLoomie1 || !containsLoomie2 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You should own both loomies to fuse them"})
		return
	}

	// Check both loomies are of the same type
	loomiesDocs, err := models.GetLoomiesByIds([]primitive.ObjectID{firstLoomieMongoId, secondLoomieMongoId}, userMongoId)

	if err != nil || len(loomiesDocs) != 2 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Error getting the loomies. Please try again later."})
		return
	}

	if loomiesDocs[0].Serial != loomiesDocs[1].Serial {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Both loomies should be of the same type"})
		return
	}

	// Check the loomies are not busy
	if loomiesDocs[0].IsBusy || loomiesDocs[1].IsBusy {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": true, "message": "Both loomies should not be busy"})
		return
	}

	// Take the max stats from the loomies
	maxHp := math.Max(float64(loomiesDocs[0].Hp), float64(loomiesDocs[1].Hp))
	maxAttack := math.Max(float64(loomiesDocs[0].Attack), float64(loomiesDocs[1].Attack))
	maxDefense := math.Max(float64(loomiesDocs[0].Defense), float64(loomiesDocs[1].Defense))

	// "Fuse" the loomies (Delete one and update the other)
	var loomieToUpdate, loomieToDelete interfaces.UserLoomiesRes
	var availableExperience float64
	var minLvl int
	availableExperience = float64(loomiesDocs[0].Experience) + float64(loomiesDocs[1].Experience)

	// The loomie with the highest level will be the one that will be updated
	if loomiesDocs[0].Level > loomiesDocs[1].Level {
		loomieToUpdate = loomiesDocs[0]
		loomieToDelete = loomiesDocs[1]
		minLvl = loomiesDocs[1].Level
	} else {
		loomieToUpdate = loomiesDocs[1]
		loomieToDelete = loomiesDocs[0]
		minLvl = loomiesDocs[0].Level
	}

	// Increment the available experience by 120% of the experience of the loomie with the lowest level
	availableExperience += utils.GetRequiredExperience(minLvl) * 1.2

	// We reset the Loomie experience because that experience is already considered in the availableExperience variable
	loomieToUpdate.Experience = 0
	var experienceToAdd, neededExperienceToNextLevel float64

	// Check if the loomie has leveled up
	for (loomieToUpdate.Experience + availableExperience) >= utils.GetRequiredExperience(loomieToUpdate.Level+1) {
		neededExperienceToNextLevel = utils.GetRequiredExperience(loomieToUpdate.Level + 1)
		experienceToAdd = math.Min(availableExperience, neededExperienceToNextLevel)
		experienceToAdd = utils.FixeFloat(experienceToAdd, 4)
		loomieToUpdate.Level++
		loomieToUpdate.Experience = 0
		availableExperience -= experienceToAdd
	}

	// Add the remaining experience to the loomie
	loomieToUpdate.Experience = utils.FixeFloat(availableExperience, 4)

	// Update the loomie
	loomieToUpdate.Hp = int(maxHp)
	loomieToUpdate.Attack = int(maxAttack)
	loomieToUpdate.Defense = int(maxDefense)
	err = models.FuseLoomies(userMongoId, loomieToUpdate, loomieToDelete)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Error fusing the loomies. Please try again later."})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Loomies fused successfully",
	})
}

// HandleCaptureLoomie Handle the request to capture a wild loomie with a loomball
func HandleCaptureLoomie(c *gin.Context) {
	// Get the Loomies Id, Latitude, Longitude and LoomieBall Id from the request body
	loomie_req := interfaces.CatchLoomieForm{}
	// Get the User Id
	userid, _ := c.Get("userid")

	if err := c.BindJSON(&loomie_req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Bad request"})
		return
	}

	//Check if fild is empty
	if loomie_req.LoomieId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "No Loomie ID was provided"})
		return
	}

	//Check and get the wild loomie
	loomie, err := models.GetWildLoomieById(loomie_req.LoomieId)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": true, "message": "Loomie was not found"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
			return
		}
	}

	//Check if the wild loomie is near
	isNear := utils.IsNear(interfaces.Coordinates{
		Latitude:  loomie.Latitude,
		Longitude: loomie.Longitude,
	}, interfaces.Coordinates{
		Latitude:  loomie_req.Latitude,
		Longitude: loomie_req.Longitude,
	})

	if !isNear {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "User is not near the loomie"})
		return
	}

	//Check and get user
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

	//Check if user id alrady exists in array UsersAlreadyCapturedIt from wild loomie
	insertUserInWildLoomie := models.CheckIfUserInArrayOfWildLoomie(loomie, user)

	fmt.Println(insertUserInWildLoomie)

	if !insertUserInWildLoomie {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "User already caught this loomie"})
		return
	}

	//Transform from string to primitive Object ID
	mongoid, err := primitive.ObjectIDFromHex(loomie_req.LoomballId)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "The given Loomball id is not valid"})
		return
	}

	//Remove the loomBall from inventory
	err = models.DecrementItemFromUserInventory(user.Id, mongoid, 1)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": true, "message": "The given loomball was not found"})
		return
	}

	array_ball := []primitive.ObjectID{}

	array_ball = append(array_ball, mongoid)

	//Get loomBall
	loomball, err := models.GetLoomballsFromIds(array_ball)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
		return
	}

	//Check if the loomie was caught
	was_captured := models.WasSuccessfulCapture(loomie, loomball[0])

	if was_captured {
		//Insert user id in array UsersAlreadyCapturedIt from wild loomie
		err = models.InsertUserInArrayOfWildLoomie(loomie, user)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "User already caught this loomie"})
			return
		}

		caught_loomies := interfaces.CaughtLoomie{Owner: user.Id,
			IsBusy:     false,
			Serial:     loomie.Serial,
			Name:       loomie.Name,
			Types:      loomie.Types,
			Rarity:     loomie.Rarity,
			HP:         loomie.HP,
			Attack:     loomie.Attack,
			Defense:    loomie.Defense,
			Level:      loomie.Level,
			Experience: loomie.Experience,
		}

		//Save the wild loomie in the caught roomie collection
		caught_loomies_id, err := models.InsertInCaughtLoomies(caught_loomies)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
			return
		}

		//Save the wild loomie on the user
		err = models.AddToUserLoomies(user, caught_loomies_id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
			return
		}

		c.IndentedJSON(http.StatusOK, gin.H{
			"error":        false,
			"was_captured": was_captured,
			"message":      "The loomie was captured",
		})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"error":        false,
		"was_captured": was_captured,
		"message":      "The Loomie was not caught. Try again!",
	})
}
