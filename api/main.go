package main

import (
	"github.com/PedroChaparro/loomies-backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Setup server
	// Adding an useless comment
	engine := gin.Default()
	routes.SetupRoutes(engine)
	engine.Run()
}
