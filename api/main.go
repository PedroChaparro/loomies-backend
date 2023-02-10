package main

import (
	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/gin-gonic/gin"
)

func main() {
	// Prepare environment
	configuration.Load()

	// Start gin server
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run()
}
