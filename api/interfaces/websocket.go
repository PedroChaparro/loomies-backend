package interfaces

import "github.com/gorilla/websocket"

// WsClient stores the websocket connection to be able
// to send messages to the client
type WsClient struct {
	Connection *websocket.Conn
	// The channel will be used to internally send messages
	// and contol the active connections
	Channel chan<- string
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
