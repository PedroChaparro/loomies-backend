package routes

import (
	"github.com/PedroChaparro/loomies-backend/controllers"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(engine *gin.Engine) {
	// User
	engine.POST("/signup", controllers.HandleSignUp)
	engine.POST("/code_validation", controllers.HandleCodeValidation)
	engine.POST("/new_code", controllers.HandleNewCodeValidation)
	engine.GET("/user/loomies", middlewares.MustProvideAccessToken(), controllers.HandleGetLoomies)

	// Session
	engine.POST("/login", controllers.HandleLogIn)
	engine.GET("/whoami", middlewares.MustProvideAccessToken(), controllers.HandleWhoami)
	engine.GET("/refresh", middlewares.MustProvideRefreshToken(), controllers.HandleRefresh)

	// Gyms
	engine.POST("/near_gyms", middlewares.MustProvideAccessToken(), controllers.HandleNearGyms)
	engine.POST("/gyms/claim_reward", middlewares.MustProvideAccessToken(), controllers.HandleClaimReward)

	// Loomies
	engine.POST("/near_loomies", middlewares.MustProvideAccessToken(), controllers.HandleNearLoomies)

	//Items
	engine.GET("/items", middlewares.MustProvideAccessToken(), controllers.HandleGetItems)
}
