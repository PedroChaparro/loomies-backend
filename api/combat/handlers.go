package combat

import (
	"fmt"
	"math"
	"time"

	"github.com/PedroChaparro/loomies-backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ######################### Combat handlers #########################
// This functions cannot be defined in the controllers package
// because in that case, the handlers cannot access the Ws* structs
// due to the circular dependency between the combat and controllers
// packages (controllers package imports combat to use the types and
// combat imports controllers to use the handlers)

// handleSendAttack handles the "GYM_ATTACK" message type to send an attack to the player
func handleSendAttack(combat *WsCombat) {
	// Check if the types of the loomie were obtained before
	gymLoomie := combat.CurrentGymLoomie
	playerLoomie := combat.CurrentPlayerLoomie

	// For each type (Currently, there is only one or two types per loomie)
	for _, value := range gymLoomie.Types {
		// Check if the type was cached before
		_, cached := GlobalWsHub.CachedStrongAgainst[value]

		// If the type was not obtained before, get it from the database
		if !cached {
			typeDetails, err := models.GetLoomieTypeDetailsByName(value)

			if err != nil {
				combat.SendMessage(WsMessage{
					Type:    "ERROR",
					Message: "Error getting the loomie type details",
				})
				return
			}

			// Cache the type details
			GlobalWsHub.CachedStrongAgainst[value] = make([]string, 0)
			GlobalWsHub.CachedStrongAgainst[value] = typeDetails.StrongAgainst
		}
	}

	// Calculate the damage
	actualGymLoomieDamage := gymLoomie.Attack
	accumulatedDamage := actualGymLoomieDamage

TYPES_LOOP:
	for _, gymLoomieType := range gymLoomie.Types {
		for _, playerLoomieType := range playerLoomie.Types {
			for _, strongAgainst := range GlobalWsHub.CachedStrongAgainst[gymLoomieType] {
				if strongAgainst == playerLoomieType {
					accumulatedDamage *= 2
					break TYPES_LOOP
				}
			}
		}
	}

	// Apply the user loomie defense
	accumulatedDamage -= accumulatedDamage * (playerLoomie.Defense / 100)
	accumulatedDamage = int(math.Max(float64(accumulatedDamage), float64(actualGymLoomieDamage)*0.1))

	// Send the attack "notification" to the client
	combat.SendMessage(WsMessage{
		Type:    "GYM_ATTACK_CANDIDATE",
		Message: "Enemy loomie is about to attack",
	})

	// Materialize the attack after 1 second
	go func() {
		time.Sleep(1 * time.Second)

		if len(combat.Dodges) == 0 {
			combat.Dodges <- false
		}

		return
	}()

	// Just wait for the first message (dodge or not)
	var wasAttackDodged bool

	for {
		select {
		case dodged := <-combat.Dodges:
			wasAttackDodged = dodged
		}

		break
	}

	// Send the attack result to the client
	if wasAttackDodged {
		combat.SendMessage(WsMessage{
			Type:    "GYM_ATTACK_DODGED",
			Message: fmt.Sprintf("Your loomie %s dodged the attack", playerLoomie.Name),
		})

		return
	}

	playerLoomie.Hp -= accumulatedDamage

	// If the player loomie was weakened, remove it from the player loomies
	if playerLoomie.Hp <= 0 {
		weaknedLoomieId := playerLoomie.Id
		combat.PlayerLoomies = combat.PlayerLoomies[1:]

		combat.SendMessage(WsMessage{
			Type:    "USER_LOOMIE_WEAKENED",
			Message: fmt.Sprintf("Your loomie %s was weakened", playerLoomie.Name),
			Payload: map[string]interface{}{
				"loomie_id": weaknedLoomieId,
			},
		})

		if len(combat.PlayerLoomies) == 0 {
			combat.SendMessage(WsMessage{
				Type:    "USER_HAS_LOST",
				Message: "You have lost the battle. Try fusioning your loomies or caught more loomies to improve your team",
			})

			// Send the message to close all the active loops
			combat.Close <- true
			return
		}

		// Update the current player loomie
		combat.CurrentPlayerLoomie = &combat.PlayerLoomies[0]

		combat.SendMessage(WsMessage{
			Type:    "UPDATE_PLAYER_LOOMIE",
			Message: fmt.Sprintf("Your loomie %s is now your active loomie", combat.CurrentPlayerLoomie.Name),
			Payload: map[string]interface{}{
				"loomie": combat.CurrentPlayerLoomie,
			},
		})

		return
	}

	combat.SendMessage(WsMessage{
		Type:    "UPDATE_USER_LOOMIE_HP",
		Message: fmt.Sprintf("Your loomie %s received %d damage", playerLoomie.Name, accumulatedDamage),
		Payload: map[string]interface{}{
			"loomie_id": playerLoomie.Id,
			"hp":        playerLoomie.Hp,
		},
	})
}

// handleReceiveAttack handles the "USER_ATTACK" message type to receive an attack from the player
func handleReceiveAttack(combat *WsCombat) {
	gymLoomie := combat.CurrentGymLoomie
	playerLoomie := combat.CurrentPlayerLoomie

	// Check if the gym loomie was fought by the player loomie before
	_, alreadyFought := combat.FoughtGymLoomies[gymLoomie.Id]
	if !alreadyFought {
		combat.FoughtGymLoomies[gymLoomie.Id] = make([]primitive.ObjectID, 0)
		combat.FoughtGymLoomies[gymLoomie.Id] = append(combat.FoughtGymLoomies[gymLoomie.Id], playerLoomie.Id)
	}

	// Get the types
	for _, value := range playerLoomie.Types {
		// Check if the type was cached before
		_, cached := GlobalWsHub.CachedStrongAgainst[value]

		// If the type was not obtained before, get it from the database
		if !cached {
			typeDetails, err := models.GetLoomieTypeDetailsByName(value)

			if err != nil {
				combat.SendMessage(WsMessage{
					Type:    "ERROR",
					Message: "Error getting the loomie type details",
				})
				return
			}

			GlobalWsHub.CachedStrongAgainst[value] = make([]string, 0)
			GlobalWsHub.CachedStrongAgainst[value] = typeDetails.StrongAgainst
		}
	}

	// Calculate the damage
	actualPlayerLoomieDamage := playerLoomie.Attack
	accumulatedDamage := actualPlayerLoomieDamage

TYPES_LOOP:
	for _, playerLoomieType := range playerLoomie.Types {
		for _, gymLoomieType := range gymLoomie.Types {
			for _, strongAgainst := range GlobalWsHub.CachedStrongAgainst[playerLoomieType] {
				if strongAgainst == gymLoomieType {
					accumulatedDamage *= 2
					break TYPES_LOOP
				}
			}
		}
	}

	// Apply the user loomie defense
	accumulatedDamage -= accumulatedDamage * (gymLoomie.Defense / 100)
	accumulatedDamage = int(math.Max(float64(accumulatedDamage), float64(actualPlayerLoomieDamage)*0.1))

	// Check if the gym loomie dodged the attack
	gymLoomieDodgeProbability := 10
	luckyNumber := getRandomInt(1, 100)

	if luckyNumber <= gymLoomieDodgeProbability {
		combat.SendMessage(WsMessage{
			Type:    "USER_ATTACK_DODGED",
			Message: fmt.Sprintf("Enemy loomie %s dodged the attack", gymLoomie.Name),
		})

		return
	}

	gymLoomie.Hp -= accumulatedDamage

	// If the gym loomie was weakened, remove it from the gym loomies
	if gymLoomie.Hp <= 0 {
		wenakenedLoomieId := gymLoomie.Id
		combat.GymLoomies = combat.GymLoomies[1:]

		combat.SendMessage(WsMessage{
			Type:    "GYM_LOOMIE_WEAKENED",
			Message: fmt.Sprintf("Enemy loomie %s was weakened", gymLoomie.Name),
			Payload: map[string]interface{}{
				"loomie_id": wenakenedLoomieId,
			},
		})

		combat.DefeatedGymLoomies = append(combat.DefeatedGymLoomies, wenakenedLoomieId)

		if len(combat.GymLoomies) == 0 {
			combat.SendMessage(WsMessage{
				Type:    "USER_HAS_WON",
				Message: "You have won the battle. Now you own this gym",
			})

			combat.Close <- true
			return
		}

		// Update the current gym loomie
		combat.CurrentGymLoomie = &combat.GymLoomies[0]

		combat.SendMessage(WsMessage{
			Type:    "UPDATE_GYM_LOOMIE",
			Message: fmt.Sprintf("Enemy loomie %s is now the gym's active loomie", combat.CurrentGymLoomie.Name),
			Payload: map[string]interface{}{
				"loomie": combat.CurrentGymLoomie,
			},
		})

		return
	}

	combat.SendMessage(WsMessage{
		Type:    "UPDATE_GYM_LOOMIE_HP",
		Message: fmt.Sprintf("Enemy loomie %s received %d damage", gymLoomie.Name, accumulatedDamage),
		Payload: map[string]interface{}{
			"loomie_id": gymLoomie.Id,
			"hp":        gymLoomie.Hp,
		},
	})
}

// handleClearDodgeChannel Clears the dodge channel to avoid collisions between attacks
func handleClearDodgeChannel(combat *WsCombat) {
	for len(combat.Dodges) > 0 {
		<-combat.Dodges
	}
}
