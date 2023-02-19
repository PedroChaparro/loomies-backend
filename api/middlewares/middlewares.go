package middlewares

import (
	"net/http"

	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
)

// MustProvideAccessToken checks if a valid access token was provided in the header
func MustProvideAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get access token from header
		accessToken := c.GetHeader("Access-Token")

		if accessToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Access token is required"})
			return
		}

		// Check if access token is valid
		id, error := utils.ValidateAccessToken(accessToken)
		if error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": error.Error()})
			return
		}

		// Set user id to context
		c.Set("userid", id)
	}
}

// MustProvideRefreshToken checks if a valid refresh token was provided in the header
func MustProvideRefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get refresh token from header
		refreshToken := c.GetHeader("Refresh-Token")

		if refreshToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Refresh token is required"})
			return
		}

		// Check if refresh token is valid
		id, error := utils.ValidateRefreshToken(refreshToken)
		if error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": error.Error()})
			return
		}

		// Set user id to context
		c.Set("userid", id)
	}
}
