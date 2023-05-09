package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PedroChaparro/loomies-backend/combat"
	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// The upgrader is used to upgrade the http connection to a websocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Initialy, this accept all the origins
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleCombatRegister Handles the request to register a combat returning a token to authenticate the user with the websocket endpoint
func HandleCombatRegister(c *gin.Context) {
	// Receive the request body
	var payload interfaces.RegisterCombatReq
	if err := c.ShouldBindJSON(&payload); err != nil || payload.GymID == "" || payload.Latitude == 0 || payload.Longitude == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "The request body should contain the gym id and the user coordinates"})
		return
	}

	// Check the gym is near the user coordinates
	gymDoc, err := models.GetGymFromID(payload.GymID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to get the gym. Please try again later."})
		return
	}

	// Check the gym is near the user coordinates
	if !utils.IsNear(interfaces.Coordinates{
		Latitude:  gymDoc.Latitude,
		Longitude: gymDoc.Longitude,
	}, interfaces.Coordinates{
		Latitude:  payload.Latitude,
		Longitude: payload.Longitude,
	}) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You are too far away from the gym"})
		return
	}

	// Get the user and the gym from the database
	userID, _ := c.Get("userid")
	userMongoID, _ := primitive.ObjectIDFromHex(userID.(string))
	userDoc, _ := models.GetUserById(userID.(string))
	gymDoc, _ = models.GetGymFromID(payload.GymID)

	// Check the user is not the gym owner
	if userDoc.Id == gymDoc.Owner {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You can't challenge your own gym"})
		return
	}

	// Check the user and the gym have a loomie team
	if len(userDoc.LoomieTeam) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You must have at least one loomie in your team to start a combat."})
		return
	}

	if len(gymDoc.Protectors) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "The gym doesn't have any protector loomies. Try with another gym."})
		return
	}

	// Check the user is not in combat
	_, err = models.GetActiveCombatByUseId(userMongoID)

	if err == nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": true, "message": "You are already in combat"})
		return
	}

	if err != mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to get the active combat. Please try again later."})
		return
	}

	// Check the user has not challenged the gym recently
	lastUserChallenge, err := models.GetLastGymChallengeTimestamp(gymDoc.Id, userMongoID)
	gymsChallengesTimeout := configuration.GetCombatChallengeTimeout()
	previousAttackTime := time.Unix(lastUserChallenge.Timestamp, 0)
	nextValidChallenge := previousAttackTime.Add(time.Duration(gymsChallengesTimeout) * time.Minute)

	if err != nil && err != mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to get the last gym challenge. Please try again later."})
		return
	}

	if time.Now().Before(nextValidChallenge) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You have already challenged this gym recently. Please try again later"})
		return
	}

	// Create a token to authenticate the user with the websocket endpoint
	token, err := utils.CreateWsToken(userID.(string), payload.GymID, payload.Latitude, payload.Longitude)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to craete a token for the combat. Please try again later."})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"error": false, "message": "Token was created successfully", "combat_token": token})
}

// HandleCombatInit Handles the request to initialize a combat from a combat token returning the websocket connection
func HandleCombatInit(c *gin.Context) {
	// Receive the token from the params
	token := c.Query("token")

	if token == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "The token is required"})
		return
	}

	// Validate the token and get the claims
	claims, err := utils.ValidateWsToken(token)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": true, "message": "The token is invalid"})
		return
	}

	// Get the gym from the database
	gymDoc, err := models.GetGymFromID(claims.GymID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to get the gym. Please try again later."})
		return
	}

	// Check the gym is not already in combat
	hub := combat.GlobalWsHub
	inCombat := hub.Includes(claims.GymID)

	if inCombat {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": true, "message": "The gym is already in combat"})
		return
	}

	// Get the user and gym loomies
	var userCombatLoomies, gymCombatLoomies []interfaces.CombatLoomie
	user, _ := models.GetUserById(claims.UserID)
	userLoomies, _ := models.GetLoomiesByIds(user.LoomieTeam, user.Id)
	gymLoomies, _ := models.GetLoomiesByIds(gymDoc.Protectors, primitive.NilObjectID)

	// Uncomment this to see the user and gym loomies
	// NOTE: This can be removed in further pull requests

	/*
		fmt.Println("User loomies:")
		for _, loomie := range userLoomies {
			fmt.Printf("%+v\n", loomie)
		}

		fmt.Println("Gym loomies:")
		for _, loomie := range gymLoomies {
			fmt.Printf("%+v\n", loomie)
		}
	*/

	if len(userLoomies) == 0 || len(gymLoomies) == 0 {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": true, "message": "You or the gym doesn't have loomies. Please catch some loomies or search for a gym with loomies."})
		return
	}

	// Upgrade the connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to upgrade the connection to a websocket connection"})
		return
	}

	// Update the loomies stats
	for _, loomie := range userLoomies {
		userCombatLoomies = append(userCombatLoomies, *loomie.ToCombatLoomie())
	}

	for _, loomie := range gymLoomies {
		gymCombatLoomies = append(gymCombatLoomies, *loomie.ToCombatLoomie())
	}

	Combat := &combat.WsCombat{
		PlayerID:                 user.Id,
		GymID:                    claims.GymID,
		Connection:               conn,
		LastMessageTimestamp:     time.Now().Unix(),
		NextValidAttackTimestamp: 0,
		PlayerLoomies:            userCombatLoomies,
		AlivePlayerLoomies:       len(userCombatLoomies),
		GymLoomies:               gymCombatLoomies,
		AliveGymLoomies:          len(gymCombatLoomies),
		CurrentGymLoomie:         &gymCombatLoomies[0],
		CurrentPlayerLoomie:      &userCombatLoomies[0],
		FoughtGymLoomies:         make(map[primitive.ObjectID][]*interfaces.CombatLoomie),
		Dodges:                   make(chan bool, 1),
		Close:                    make(chan bool, 1),
	}

	// Update the last user challenge
	err = models.UpdateLastGymChallengeTimestamp(gymDoc.Id, user.Id)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to update the last gym challenge. Please try again later."})
		return
	}

	// Register the connection on the hub
	hub.Register(claims.GymID, Combat)

	// Send the initial loomies to the client
	Combat.SendMessage(combat.WsMessage{
		Type:    "start",
		Message: "The combat has started.",
		Payload: gin.H{
			"player_loomie":      Combat.CurrentPlayerLoomie,
			"alive_user_loomies": Combat.AlivePlayerLoomies,
			"gym_loomie":         Combat.CurrentGymLoomie,
			"alive_gym_loomies":  Combat.AliveGymLoomies,
		},
	})

	// Listen for messages
	Combat.Listen(hub)
	// NOTE: The response is sended automatically when upgrading the connection
}
