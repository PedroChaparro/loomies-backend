package controllers

import (
	"net/http"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
)

// HandleNearGyms Handle the request to obtain the gyms near the user coordinates
func HandleNearGyms(c *gin.Context) {
	bodyCoord := interfaces.Coordinates{}

	if err := c.BindJSON(&bodyCoord); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": true, "message": "JSON payload is invalid or missing"})
		return
	}

	nearGyms, err := models.GetNearGyms(bodyCoord.Latitude, bodyCoord.Longitude)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Internal server error"})
		return
	}

	if len(nearGyms) == 0 {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": true, "message": "Gyms Not Found",
			"nearGyms": []interfaces.Gym{},
		})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"error": false, "message": "Gyms have been found in near areas",
			"nearGyms": nearGyms,
		})
	}
}
