package controllers

import (
	"net/http"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
)

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
	// Receive the gym id and the token from the params
}
