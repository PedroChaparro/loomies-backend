package combat

import (
	"encoding/json"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/gorilla/websocket"
)

// WsCombat stores the websocket connection to be able
// to send messages to the client
type WsCombat struct {
	// We keep the gym id to easily remove it from the map
	// when the connection is closed
	GymID string
	// The connecton to exchange messages with the client
	Connection *websocket.Conn
	// We keep track of the last message timestamp to finish
	// the connection if the client is inactive for too long
	LastMessageTimestamp int64
	// Loomie teams in combat
	GymLoomies    []interfaces.UserLoomiesRes
	PlayerLoomies []interfaces.UserLoomiesRes
	// Current loomie in combat
	CurrentGymLoomie    *interfaces.UserLoomiesRes
	CurrentPlayerLoomie *interfaces.UserLoomiesRes
}

// WsMessage is the message that is sent to the client
type WsMessage struct {
	// Type represent the possible actions the player can do
	// Eg. "Attack", "Change current loomie", etc.
	Type    string `json:"type"`
	Message string `json:"message"`
	// The payload field allows to send any kind of data in JSON format
	Payload interface{} `json:"payload,omitempty"`
}

// WsHub is the hub that stores all the clients
type WsHub struct {
	// The key of the map is the Gym id, so, there can only
	// be one client per gym
	Combats map[string]*WsCombat
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
func (hub *WsHub) Register(gym string, combat *WsCombat) bool {
	if hub.Includes(gym) {
		return false
	}

	hub.Combats[gym] = combat
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

// SendMessage sends a message to the client
func (combat *WsCombat) SendMessage(message WsMessage) {
	jsonMessage, _ := json.Marshal(message)
	stringJson := string(jsonMessage)
	combat.Connection.WriteJSON(stringJson)
}

// Listen is the function that listens for messages from the client
func (combat *WsCombat) Listen(hub *WsHub) {
	// --- Close the connection when the function ends ---
	defer func() {
		hub.Unregister(combat.GymID)
		// TODO: This should remove the combat from the database
	}()

	// --- Independent goroutine to check if the client is inactive ---
	go func() {
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				if time.Now().Unix()-combat.LastMessageTimestamp > 30 {
					combat.Connection.Close()
					return
				}
			}
		}
	}()

	// --- Endless loop to listen for messages ---
	for {
		_, message, err := combat.Connection.ReadMessage()

		// If there is an error, is probably because the connection
		// was closes, so, we exit the loop
		if err != nil {
			return
		}

		// Parse message to JSON
		var wsMessage WsMessage
		err = json.Unmarshal(message, &wsMessage)

		// Check the message type and send to the corresponding handler
		switch wsMessage.Type {
		case "greeting":
			handleGreetingMessageType(combat)
		}
	}
}
