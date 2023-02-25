package routes

import (
	"github.com/PedroChaparro/loomies-backend/controllers"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(engine *gin.Engine) {
	engine.POST("/signup", controllers.HandleSignUp)
	engine.POST("/login", controllers.HandleLogIn)
	engine.GET("/whoami", middlewares.MustProvideAccessToken(), controllers.HandleWhoami)
	engine.GET("/refresh", middlewares.MustProvideRefreshToken(), controllers.HandleRefresh)
	engine.POST("/near_gyms", middlewares.MustProvideAccessToken(), controllers.HandleNearGyms)
}
