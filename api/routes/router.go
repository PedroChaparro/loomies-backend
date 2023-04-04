package routes

import (
	"github.com/PedroChaparro/loomies-backend/controllers"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(engine *gin.Engine) {
	// User
	engine.GET("/user/loomies", middlewares.MustProvideAccessToken(), controllers.HandleGetLoomies)
	engine.GET("/user/loomie-team", middlewares.MustProvideAccessToken(), controllers.HandleGetLoomieTeam)
	engine.POST("/user/password/code", controllers.HandleResetPasswordCodeRequest)
	engine.PUT("/user/password", controllers.HandleResetPassword)
	engine.POST("/user/signup", controllers.HandleSignUp)
	engine.POST("/user/validate/code", controllers.HandleAccountValidationCodeRequest)
	engine.POST("/user/validate", controllers.HandleAccountValidation)

	// Session
	engine.POST("/session/login", controllers.HandleLogIn)
	engine.GET("/session/whoami", middlewares.MustProvideAccessToken(), controllers.HandleWhoami)
	engine.GET("/session/refresh", middlewares.MustProvideRefreshToken(), controllers.HandleRefresh)

	// Gyms
	engine.POST("/gyms/near", middlewares.MustProvideAccessToken(), controllers.HandleNearGyms)
	engine.POST("/gyms/claim-reward", middlewares.MustProvideAccessToken(), controllers.HandleClaimReward)

	// Loomies
	engine.POST("/loomies/near", middlewares.MustProvideAccessToken(), controllers.HandleNearLoomies)
	engine.GET("/loomies/exists/:id", middlewares.MustProvideAccessToken(), controllers.HandleValidateLoomieExists)
	engine.POST("/loomies/fuse", middlewares.MustProvideAccessToken(), controllers.HandleFuseLoomies)
	engine.POST("/loomies/capture", middlewares.MustProvideAccessToken(), controllers.HandleCaptureLoomie)

	// Items
	engine.GET("/user/items", middlewares.MustProvideAccessToken(), controllers.HandleGetItems)
}
