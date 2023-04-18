package main

import (
	"github.com/PedroChaparro/loomies-backend/combat"
	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Setup server and default routes
	engine := gin.Default()
	routes.SetupRoutes(engine)

	// Set gin mode to release if in production
	if configuration.GetRunningEnvironment() == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup websocket routes
	hub := combat.WsHub{
		Combats:             make(map[string]*combat.WsCombat),
		CachedStrongAgainst: make(map[string][]string),
	}

	combat.GlobalWsHub = &hub
	routes.SetupWebSocketRoutes(engine)

	// Start the server
	engine.Run()
}
