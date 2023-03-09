package main

import (
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create an unique instance of the Web Socket hub
	// and initialize the Clients to an empty map
	wsHub := interfaces.WsHub{
		Clients: make(map[string]*interfaces.WsClient),
	}

	// Setup the API routes
	engine := gin.Default()
	routes.SetupRoutes(engine)
	routes.SetupWsRoutes(engine, &wsHub) // <--- Web socket routes
	engine.Run()
}
