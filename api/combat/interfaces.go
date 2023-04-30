package combat

import (
	"encoding/json"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WsCombat stores the websocket connection to be able
// to send messages to the client
type WsCombat struct {
	PlayerID primitive.ObjectID
	// We keep the gym id to easily remove it from the map when the combat ends
	GymID string
	// The connecton to exchange messages with the client
	Connection *websocket.Conn
	// Keep track of the last message timestamp to finish the combat if the client is "akf"
	LastMessageTimestamp int64
	// Keep track of the last attack timestamp to avoid spamming
	LastUserAttackTimestamp int64
	// Keep track of the alive user and gym loomies
	AlivePlayerLoomies int
	AliveGymLoomies    int
	// Loomie teams in combat
	GymLoomies    []interfaces.CombatLoomie
	PlayerLoomies []interfaces.CombatLoomie
	// Current loomie in combat
	CurrentGymLoomie    *interfaces.CombatLoomie
	CurrentPlayerLoomie *interfaces.CombatLoomie
	// Defeated loomies in combat
	FoughtGymLoomies map[primitive.ObjectID][]*interfaces.CombatLoomie
	// Channels to communicate between the goroutines
	Dodges chan bool
	Close  chan bool
}

// WsMessage is the message that is sent to the client
type WsMessage struct {
	// Type represent the possible actions the player can do
	// Eg. "Attack", "Change current loomie", etc.
	Type    string `json:"type"`
	Message string `json:"message"`
	// The payload field allows to send any kind of data in JSON format
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// WsHub is the hub that stores all the clients
type WsHub struct {
	// The key of the map is the Gym id, so, there can only
	// be one client per gym
	Combats map[string]*WsCombat
	// Map to store the strong against types
	CachedStrongAgainst map[string][]string
}

// GlobalWsHub is the global hub that stores all the clients
// This is initialized in the main.go file
var GlobalWsHub *WsHub

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

// UpdatedLastReceivedMessageTimestamp updates the timestamp of the last message received from the client
func (combat *WsCombat) UpdatedLastReceivedMessageTimestamp() {
	combat.LastMessageTimestamp = time.Now().Unix()
}

// UpdatedLastUserAttackTimestamp updates the timestamp of the last attack sent by the client
func (combat *WsCombat) UpdatedLastUserAttackTimestamp() {
	combat.LastUserAttackTimestamp = time.Now().Unix()
}

// SendMessage sends a message to the client
func (combat *WsCombat) SendMessage(message WsMessage) {
	combat.Connection.WriteJSON(message)
}

// Listen is the function that listens for messages from the client
func (combat *WsCombat) Listen(hub *WsHub) {
	// --- Close the connection when the function ends ---
	defer func() {
		// Remove the combat from the hub, so the gym can be challenged again
		hub.Unregister(combat.GymID)

		// Mark the combat as closed
		gymIdMongo, _ := primitive.ObjectIDFromHex(combat.GymID)
		models.FinishGymChallenge(gymIdMongo, combat.PlayerID)
	}()

	// --- Independent goroutine to check if the client is inactive ---
	go func() {
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-combat.Close:
				combat.Connection.Close()
				ticker.Stop()
				return
			case <-ticker.C:
				// If the last message received is older than 5 minutes, close the connection
				if time.Now().Unix()-combat.LastMessageTimestamp > 300 {
					combat.SendMessage(WsMessage{
						Type:    "COMBAT_TIMEOUT",
						Message: "You have been inactive for too long, the combat has ended",
					})

					combat.Connection.Close()
					return
				}
			}
		}
	}()

	// --- Independet goroutine to send attacks from the gym to the player ---
	go func() {
		minTimeout, maxTimeout := configuration.GetCombatTimeouts()
		randomSeconds := getRandomInt(minTimeout, maxTimeout)
		ticker := time.NewTicker(time.Duration(randomSeconds) * time.Second)

		for {
			// Wait for the ticker to send a message
			select {
			case <-combat.Close:
				return
			case <-ticker.C:
				handleClearDodgeChannel(combat)
				handleSendAttack(combat)
			}

			// Reset the ticker and pick a new random interval
			ticker.Stop()
			randomSeconds := getRandomInt(minTimeout, maxTimeout)
			ticker = time.NewTicker(time.Duration(randomSeconds) * time.Second)
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
		case "USER_DODGE":
			if len(combat.Dodges) < 1 {
				combat.Dodges <- true
			}
			combat.UpdatedLastReceivedMessageTimestamp()

		case "USER_ATTACK":
			handleReceiveAttack(combat)
			combat.UpdatedLastReceivedMessageTimestamp()

		case "USER_USE_ITEM":
			handleUseItem(combat, wsMessage)
			combat.UpdatedLastReceivedMessageTimestamp()

		case "USER_ESCAPE_COMBAT":
			handleEscapeCombat(combat)

		case "USER_CHANGE_LOOMIE":
			handleChangeLoomie(combat, wsMessage)
			combat.UpdatedLastReceivedMessageTimestamp()

		case "USER_GET_LOOMIE_TEAM":
			handleGetUserTeam(combat, wsMessage)
			combat.UpdatedLastReceivedMessageTimestamp()
		}
	}
}
