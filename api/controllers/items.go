package controllers

import (
	"net/http"

	"github.com/PedroChaparro/loomies-backend/interfaces"
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

	user_items := []interfaces.GetItem{}
	
	for i:= 0; i < len(user.Items); i++{
		id := user.Items[i].Id
		item, err := models.GetItemById(id)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Item was not found"})
				return
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
				return
			}
		}
		data := interfaces.GetItem{Id: item.Id, Name: item.Name, Description: item.Description, Target: item.Target, Is_combat_item:item.Is_combat_item, Quantity:user.Items[i].Quantity}
		user_items = append(user_items, data)
	}
	
	c.IndentedJSON(http.StatusOK, gin.H{
		"items": user_items,
	})
}
