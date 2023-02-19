package controllers

import (
	"fmt"
	"net/http"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
)

func HandleNearGyms(c *gin.Context) {

	bodyCoord := interfaces.Coordinates{}

	if err := c.BindJSON(&bodyCoord); err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	nearGyms, err := models.GetNearGyms(bodyCoord.Latitude, bodyCoord.Longitude)

	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Gyms in near areas",
		"nearGyms": nearGyms,
	})
}
