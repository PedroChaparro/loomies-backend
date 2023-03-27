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

	// Setup websocket routes
	hub := combat.WsHub{
		Combats: make(map[string]*combat.WsCombat),
	}

	configuration.Globals.WsHub = &hub
	routes.SetupWebSocketRoutes(engine)

	// Start the server
	engine.Run()
}
