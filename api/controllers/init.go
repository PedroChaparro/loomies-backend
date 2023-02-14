package controllers

import (
	"fmt"

	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
)

func HandleZonesGet(c *gin.Context) {
	zones := models.GetZones()

	c.JSON(200, gin.H{
		"zones": zones,
	})
}

func HandleNearZonesGet(c *gin.Context) {
	// temporary example
	zone, _ := models.GetCurrentZone(-73.07220400291057, 7.038141994916382)
	fmt.Println(zone)

	/* c.JSON(200, gin.H{
		"zone": zone,
	}) */
}
