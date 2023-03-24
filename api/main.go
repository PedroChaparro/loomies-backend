package main

import (
	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Setup server and default routes
	engine := gin.Default()
	routes.SetupRoutes(engine)

	// Setup websocket routes
	hub := interfaces.WsHub{
		Combats: make(map[string]*interfaces.WsCombat),
	}

	configuration.Globals.WsHub = &hub
	routes.SetupWebSocketRoutes(engine, &hub)

	// Start the server
	engine.Run()
}
