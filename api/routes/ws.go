package routes

import (
	"fmt"
	"net/http"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by default
		return true
	},
}

func SetupWsRoutes(engine *gin.Engine, wsHub *interfaces.WsHub) {
	engine.GET("/api/ws/combat", func(c *gin.Context) {
		// Get the gym_id and user_id from the query string because
		// the body is not available for WS requests
		gymID := c.Query("gym_id")
		userID := c.Query("user_id")

		if gymID == "" || userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "Sould contain the gym_id and user_id"})
			return
		}

		// Verify that the gym is not already in the hub
		if wsHub.ContainsGym(gymID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "Gym is already in combat"})
			return
		}

		// If all was OK, upgrade protocol from HTTP to WS
		fmt.Printf("INFO: Upgrading protocol for gym %s and user %s... \n", gymID, userID)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		client := interfaces.WsClient{
			Connection: conn,
		}
		wsHub.AddClient(gymID, &client)
		client.ReadMessages()

		if err != nil {
			fmt.Printf("Error upgrading protocol: %v \n", err)
			return
		}

		fmt.Printf("INFO: Protocol upgraded for gym %s and user %s... \n", gymID, userID)

		// NOTE: It's not necessary to send an HTTP response here
		// because the upgrader.Upgrade() method already writes
		// the HTTP response to the client.
	})
}
