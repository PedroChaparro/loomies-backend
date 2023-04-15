package controllers

import (
	"net/http"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	c.JSON(http.StatusOK, gin.H{"error": false, "message": "User items were successfully retreived", "items": items, "loomballs": loomballs})
}

// HandleUseItem Handle the request to use and item from the inventory
func HandleUseItem(c *gin.Context) {
	var req interfaces.UseNotCombatItemReq

	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "Bad request. A JSON is needed"})
		return
	}

	if req.LoomieId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "A Loomie is required"})
		return
	}

	if req.ItemId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "An Item is required"})
		return
	}

	userid, _ := c.Get("userid")
	user, _ := primitive.ObjectIDFromHex(userid.(string))

	loomieId, _ := primitive.ObjectIDFromHex(req.LoomieId)
	itemId, err := primitive.ObjectIDFromHex(req.ItemId)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "The given item id is not valid"})
		return
	}

	var item interfaces.PopulatedInventoryItem

	item, err = models.GetItemFromUserInventory(user, itemId, true)

	if err != nil {
		if err.Error() == "USER_DOES_NOT_OWN_ITEM" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "You don't own the given item"})
			return
		}

		if err.Error() == "ITEM_NOT_FOUND" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": true, "message": "The given item was not found or is a combat item"})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error getting item from user"})
		return
	}

	// -------------------------
	// This is a temporary solution but works becase right now we only support one item
	// that can be used without being in combat
	if !(item.Serial == 7) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "The given item can only be used in combat"})
		return
	}

	if !(item.Quantity > 0) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "User doesn't have enough items"})
		return
	}

	err = models.DecrementItemFromUserInventory(user, itemId, 1)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error decrementing item from user"})
		return
	}

	_, err = models.IncrementLoomieLevel(user, loomieId, 1)

	if err != nil {
		// Item quantity is restored to the user
		models.IncrementItemFromUserInventory(user, itemId, 1)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error updating level of Loomie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": false, "message": "Item was successfully used"})
}
