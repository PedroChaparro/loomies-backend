package routes

import (
	"github.com/PedroChaparro/loomies-backend/controllers"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/gin-gonic/gin"
)

// --- Websocket Routes ---
func SetupWebSocketRoutes(engine *gin.Engine) {
	engine.POST("/combat/register", middlewares.MustProvideAccessToken(), controllers.HandleCombatRegister)
	engine.GET("/combat", controllers.HandleCombatInit)
}
