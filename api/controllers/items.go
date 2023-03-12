package controllers

import (
	"net/http"

	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleGetItems(c *gin.Context) {
	userid, _ := c.Get("userid")
	user, err := models.GetUserById(userid.(string))

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "User was not found"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return
		}
	}

	items, err := models.GetItemById(user.Items)

	c.IndentedJSON(http.StatusOK, gin.H{
		"items": items,
	})
}
