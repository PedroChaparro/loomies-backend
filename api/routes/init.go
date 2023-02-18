package routes

import (
	"github.com/PedroChaparro/loomies-backend/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(engine *gin.Engine) {
	engine.POST("/signup", controllers.HandleSignUp)
	engine.POST("/nearzones", controllers.HandleNearGyms)
}
