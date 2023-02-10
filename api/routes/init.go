package routes

import (
	"github.com/PedroChaparro/loomies-backend/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(engine *gin.Engine) {
	engine.GET("/zones", controllers.HandleZonesGet)
}
