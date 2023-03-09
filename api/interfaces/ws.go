package interfaces

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WsClient is a struct that contains the connection to the client to be able
// to send messages through the socket.
type WsClient struct {
	// The connection to the client
	Connection *websocket.Conn
}

// WsMessage is a struct to make it possible to send  and receive messages
// through the socket.
type WsMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// WsHub is a struct that contains all the clients connected to the server.
// The key of the map is the Gym ID, this way, there can only be one player
// fighting for a gym at a time.
type WsHub struct {
	Clients map[string]*WsClient
}

// ContainsGym checks if the gym is already in the hub.
func (WsHub *WsHub) ContainsGym(gymID string) bool {
	fmt.Printf("INFO: Checking if gym %s is already in the hub... \n", gymID)
	_, ok := WsHub.Clients[gymID]
	fmt.Printf("INFO: Gym %s is already in the hub: %t \n", gymID, ok)
	return ok
}

// AddClient adds a new client to the hub.
func (wsHub *WsHub) AddClient(gymID string, wsClient *WsClient) {
	fmt.Printf("INFO: Adding client for gym %s... \n", gymID)
	wsHub.Clients[gymID] = wsClient
	fmt.Printf("INFO: Client for gym %s added... \n", gymID)
}

// ReadMessages listed for the messages sent by the client.
func (wsClient *WsClient) ReadMessages() {
	// Close the connection when the ws communication is over
	defer wsClient.Connection.Close()

	// Send random messages to the client
	go func() {
		for {
			randInt := rand.Intn(3-1) + 1
			time.Sleep(time.Duration(randInt) * time.Second)
			wsClient.Connection.WriteJSON(gin.H{
				"error":   false,
				"message": "Random message",
			})
		}
	}()

	for {
		// Read the message from the client
		_, message, err := wsClient.Connection.ReadMessage()

		// Parse message to JSON
		var wsMessage WsMessage
		_ = json.Unmarshal(message, &wsMessage)

		if err != nil {
			fmt.Printf("ERROR: Error reading message from client: %v \n", err)
			return
		}

		// Send a message back to the client
		switch wsMessage.Type {
		case "attack":
			wsClient.WriteMessage("The attack was received")
		default:
			wsClient.WriteMessage("The message type is not valid")
		}

		fmt.Println("The message was received...")
	}
}

// WriteMessage sends a message to the client.
func (wsClient *WsClient) WriteMessage(message string) {
	err := wsClient.Connection.WriteJSON(gin.H{
		"error":   false,
		"message": message,
	})

	if err != nil {
		fmt.Printf("ERROR: Error writing message to client: %v \n", err)
		return
	}

	fmt.Printf("INFO: Message sent to client: %s \n", message)
}
