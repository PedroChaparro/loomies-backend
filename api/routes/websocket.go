package routes

import (
	"net/http"

	"github.com/PedroChaparro/loomies-backend/controllers"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// The upgrader is used to upgrade the http connection to a websocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Initialy, this accept all the origins
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// --- Websocket Routes ---
func SetupWebSocketRoutes(engine *gin.Engine, hub *interfaces.WsHub) {
	engine.POST("/api/combat/register", middlewares.MustProvideAccessToken(), controllers.HandleCombatRegister)
}
