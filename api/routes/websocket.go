package routes

import (
	"github.com/PedroChaparro/loomies-backend/controllers"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/gin-gonic/gin"
)

// --- Websocket Routes ---
func SetupWebSocketRoutes(engine *gin.Engine, hub *interfaces.WsHub) {
	engine.POST("/api/combat/register", middlewares.MustProvideAccessToken(), controllers.HandleCombatRegister)
	engine.GET("/api/combat", controllers.HandleCombatInit)
}
