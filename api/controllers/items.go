package controllers

import (
	"net/http"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// HandleGetItems Handle the request to obtain the user's items
func HandleGetItems(c *gin.Context) {
	userid, _ := c.Get("userid")
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

	items, loomballs, err := models.GetItemById(user.Items)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
		return
	}

	// Prevent null responses
	if items == nil {
		items = []interfaces.UserItemsRes{}
	}

	if loomballs == nil {
		loomballs = []interfaces.UserLoomballsRes{}
	}

	c.JSON(http.StatusOK, gin.H{"error": false, "items": items, "loomballs": loomballs})
}
