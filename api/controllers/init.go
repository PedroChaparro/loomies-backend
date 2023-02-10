package controllers

import (
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
)

func HandleZonesGet(c *gin.Context) {
	zones := models.GetZones()

	c.JSON(200, gin.H{
		"zones": zones,
	})
}
