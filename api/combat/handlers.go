package combat

import (
	"fmt"
	"math"

	"github.com/PedroChaparro/loomies-backend/models"
)

// ######################### Combat handlers #########################
// This functions cannot be defined in the controllers package
// because in that case, the handlers cannot access the Ws* structs
// due to the circular dependency between the combat and controllers
// packages (controllers package imports combat to use the types and
// combat imports controllers to use the handlers)

// handleGreetingMessageType is an example of how to handle a message type
// NOTE: Please, remove this function in further pull request
func handleGreetingMessageType(combat *WsCombat) {
	// Do stuff... Here you can even use models functions to interact with the database
	// Send a message to the client
	combat.SendMessage(WsMessage{
		Type:    "greeting",
		Message: "Greeting message was received",
	})
}

// handleSendAttack handles the "GYM_ATTACK" message type to send an attack to the player
func handleSendAttack(combat *WsCombat) {
	// Check if the types of the loomie were obtained before
	gymLoomie := combat.CurrentGymLoomie
	playerLoomie := combat.CurrentPlayerLoomie

	// For each type (Currently, there is only one or two type per loomie)
	for _, value := range gymLoomie.Types {
		// Check if the type was cached before
		_, ok := combat.CachedStrongAgainst[value]

		// If the type was not obtained before, get it from the database
		if !ok {
			typeDetails, err := models.GetLoomieTypeDetailsByName(value)

			if err != nil {
				combat.SendMessage(WsMessage{
					Type:    "ERROR",
					Message: "Error getting the loomie type details",
				})
				return
			}

			// Cache the type details
			// Create the map entry
			combat.CachedStrongAgainst = make(map[string][]string)
			combat.CachedStrongAgainst[value] = typeDetails.StrongAgainst
		}
	}

	// Calc the damage
	actualGymLoomieDamage := gymLoomie.Attack * ((1 + 1/8) * gymLoomie.Level)
	accumulatedDamage := actualGymLoomieDamage

	for _, gymLoomieType := range gymLoomie.Types {
		for _, playerLoomieType := range playerLoomie.Types {
			for _, strongAgainst := range combat.CachedStrongAgainst[gymLoomieType] {
				if strongAgainst == playerLoomieType {
					accumulatedDamage *= 2
					goto CALC
				}
			}
		}
	}

CALC: // Label to break the loop
	// Apply the user loomie defense
	accumulatedDamage -= accumulatedDamage * (playerLoomie.Defense / 100)
	accumulatedDamage = int(math.Max(float64(accumulatedDamage), float64(actualGymLoomieDamage)*0.1))

	// Apply the damage to the player loomie
	playerLoomie.Hp -= accumulatedDamage

	// Send the attack message
	combat.SendMessage(WsMessage{
		Type:    "GYM_ATTACK",
		Message: fmt.Sprintf("Your loomie %s received %d damage", playerLoomie.Name, accumulatedDamage),
		Payload: map[string]interface{}{
			"damage": accumulatedDamage,
			"loomie": playerLoomie,
		},
	})

	// Check if the player loomie is dead
	if playerLoomie.Hp <= 0 {
		combat.SendMessage(WsMessage{
			Type:    "USER_LOOMIE_WEAKENED",
			Message: fmt.Sprintf("Your loomie %s was weakened", playerLoomie.Name),
		})
	}
}
