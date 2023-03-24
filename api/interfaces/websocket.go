package interfaces

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

// WsClient stores the websocket connection to be able
// to send messages to the client
type WsClient struct {
	// We keep the gym id to easily remove it from the map
	// when the connection is closed
	GymID      string
	Connection *websocket.Conn
	// We keep track of the last message timestamp to finish
	// the connection if the client is inactive for too long
	LastMessageTimestamp int64
}

// WsMessage is the message that is sent to the client
type WsMessage struct {
	// Type represent the possible actions the player can do
	// Eg. "Attack", "Change current loomie", etc.
	Type    string `json:"type"`
	Message string `json:"message"`
}

// WsHub is the hub that stores all the clients
type WsHub struct {
	// The key of the map is the Gym id, so, there can only
	// be one client per gym
	Combats map[string]*WsClient
}

type WsTokenClaims struct {
	UserID    string  `json:"user_id"`
	GymID     string  `json:"gym_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Includes checks if the hub already has a client for the gym
func (hub *WsHub) Includes(gym string) bool {
	_, ok := hub.Combats[gym]
	return ok
}

// Register registers a new client to the hub
func (hub *WsHub) Register(gym string, client *WsClient) bool {
	if hub.Includes(gym) {
		return false
	}

	hub.Combats[gym] = client
	return true
}

// Unregister removes a client from the hub
func (hub *WsHub) Unregister(gym string) bool {
	if !hub.Includes(gym) {
		return false
	}

	delete(hub.Combats, gym)
	return true
}

func (client *WsClient) Listen(hub *WsHub) {
	// --- Close the connection when the function ends ---
	defer func() {
		hub.Unregister(client.GymID)
		// TODO: This should remove the combat from the database
	}()

	// --- Independent goroutine to check if the client is inactive ---
	go func() {
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				if time.Now().Unix()-client.LastMessageTimestamp > 30 {
					client.Connection.Close()
					return
				}
			}
		}
	}()

	// --- Endless loop to listen for messages ---
	for {
		_, message, err := client.Connection.ReadMessage()

		// If there is an error, is probably because the connection
		// was closes, so, we exit the loop
		if err != nil {
			return
		}

		// Parse message to JSON
		var wsMessage WsMessage
		err = json.Unmarshal(message, &wsMessage)

		// Just print the message for now
		if err == nil {
			client.LastMessageTimestamp = time.Now().Unix()
			fmt.Println("Message received: ", wsMessage)
		}
	}
}
