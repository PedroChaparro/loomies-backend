package controllers

import (
	"net/http"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

// HandleCombatRegister creates a token for the user to authenticate
// with the websocket endpoint
func HandleCombatRegister(c *gin.Context) {
	// Receive the request body
	var payload interfaces.RegisterCombatReq
	if err := c.ShouldBindJSON(&payload); err != nil || payload.GymID == "" || payload.Latitude == 0 || payload.Longitude == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "The request body should contain the gym id and the user coordinates"})
		return
	}

	// Create a token to authenticate the user with the websocket endpoint
	userID, _ := c.Get("userid")
	token, err := utils.CreateWsToken(userID.(string), payload.GymID, payload.Latitude, payload.Longitude)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to craete a token for the combat. Please try again later."})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"error": false, "message": "Token was created successfully", "combat_token": token})
}

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

	// Check the gym is near the user coordinates
	gymDoc, err := models.GetGymFromID(claims.GymID)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to get the gym. Please try again later."})
		return
	}

	isNear := utils.IsNear(interfaces.Coordinates{
		Latitude:  gymDoc.Latitude,
		Longitude: gymDoc.Longitude,
	}, interfaces.Coordinates{
		Latitude:  claims.Latitude,
		Longitude: claims.Longitude,
	})

	if !isNear {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": true, "message": "You are not near the gym. Please go to the gym to start the combat."})
		return
	}

	// Check the gym is not already in combat
	hub := configuration.Globals.WsHub
	inCombat := hub.Includes(claims.GymID)

	if inCombat {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": true, "message": "The gym is already in combat"})
		return
	}

	// Get the user and gym loomies
	user, _ := models.GetUserById(claims.UserID)
	gym, _ := models.GetGymFromID(claims.GymID)
	userLoomies, _ := models.GetLoomiesByIds(user.LoomieTeam)
	gymLoomies, _ := models.GetLoomiesByIds(gym.Protectors)

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
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to upgrade the connection to a websocket connection"})
		return
	}

	connection := &interfaces.WsClient{
		GymID:                claims.GymID,
		Connection:           conn,
		LastMessageTimestamp: time.Now().Unix(),
	}

	// Initialize the combat on database
	err = models.InitializeCombat(claims)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Unable to initialize the combat. Please try again later."})
		return
	}

	// Register the connection on the hub
	hub.Register(claims.GymID, connection)
	connection.Listen(hub)

	// NOTE: The response is sended automatically when upgrading the connection
}
