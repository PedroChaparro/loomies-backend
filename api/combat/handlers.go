package combat

import (
	"fmt"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
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
	cacheTypeStrongAgainst(gymLoomie.Types, combat)

	// Calculate the damage
	calculatedAttack := calculateAttack(gymLoomie, playerLoomie)

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

	// Listen for the dodge message
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

	// Reduce the player loomie hp
	playerLoomie.BoostedHp -= calculatedAttack

	// Check if the player loomie was weakened
	if playerLoomie.BoostedHp <= 0 {
		weaknedLoomieId := playerLoomie.Id

		// Remove the loomie from the player loomies (Local array)
		if len(combat.PlayerLoomies) > 1 {
			combat.PlayerLoomies = combat.PlayerLoomies[1:]
		} else {
			combat.PlayerLoomies = make([]interfaces.CombatLoomie, 0)
		}

		// Notify the user that the loomie was weakened
		combat.SendMessage(WsMessage{
			Type:    "USER_LOOMIE_WEAKENED",
			Message: fmt.Sprintf("Your loomie %s was weakened", playerLoomie.Name),
			Payload: map[string]interface{}{
				"loomie_id": weaknedLoomieId,
			},
		})

		// Check if the player loose the battle
		if len(combat.PlayerLoomies) == 0 {
			combat.SendMessage(WsMessage{
				Type:    "USER_HAS_LOST",
				Message: "You have lost the battle. Try fusioning your loomies or caught more loomies to improve your team",
			})

			combat.Close <- true
			return
		}

		// Update the current player loomie
		combat.CurrentPlayerLoomie = &combat.PlayerLoomies[0]

		// Notify the user that the current player loomie was changed
		combat.SendMessage(WsMessage{
			Type:    "UPDATE_PLAYER_LOOMIE",
			Message: fmt.Sprintf("Your loomie %s is now your active loomie", combat.CurrentPlayerLoomie.Name),
			Payload: map[string]interface{}{
				"loomie": combat.CurrentPlayerLoomie,
			},
		})

		return
	} else {
		// Notify the user that the loomie hp was updated
		combat.SendMessage(WsMessage{
			Type:    "UPDATE_USER_LOOMIE_HP",
			Message: fmt.Sprintf("Your loomie %s received %d damage", playerLoomie.Name, calculatedAttack),
			Payload: map[string]interface{}{
				"loomie_id": playerLoomie.Id,
				"hp":        playerLoomie.BoostedHp,
			},
		})
	}
}

// handleReceiveAttack handles the "USER_ATTACK" message type to receive an attack from the player
func handleReceiveAttack(combat *WsCombat) {
	// Ignore spamming attacks
	if !time.Now().After(time.Unix(combat.LastUserAttackTimestamp, 0).Add(1 * time.Second)) {
		return
	}

	combat.LastUserAttackTimestamp = time.Now().Unix()
	gymLoomie := combat.CurrentGymLoomie
	playerLoomie := combat.CurrentPlayerLoomie
	cacheTypeStrongAgainst(playerLoomie.Types, combat)

	// Check if the gym loomie was fought by the player loomie before
	_, alreadyFought := combat.FoughtGymLoomies[gymLoomie.Id]

	if !alreadyFought {
		combat.FoughtGymLoomies[gymLoomie.Id] = make([]*interfaces.CombatLoomie, 0)
		combat.FoughtGymLoomies[gymLoomie.Id] = append(combat.FoughtGymLoomies[gymLoomie.Id], playerLoomie)
	} else {
		// Check if the player loomie already fought the gym loomie
		foughtByCurrentPlayerLoomie := false

		for _, previousPlayerLoomie := range combat.FoughtGymLoomies[gymLoomie.Id] {
			if previousPlayerLoomie.Id == playerLoomie.Id {
				foughtByCurrentPlayerLoomie = true
				break
			}
		}

		if !foughtByCurrentPlayerLoomie {
			combat.FoughtGymLoomies[gymLoomie.Id] = append(combat.FoughtGymLoomies[gymLoomie.Id], playerLoomie)
		}
	}

	// Calculate the attack
	calculatedAttack := calculateAttack(playerLoomie, gymLoomie)

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

	// Reduce the gym loomie hp
	gymLoomie.BoostedHp -= calculatedAttack

	// Check if the gym loomie was weakened
	if gymLoomie.BoostedHp <= 0 {
		wenakenedLoomieId := gymLoomie.Id

		// Remove the loomie from the gym loomies (Local array)
		if len(combat.GymLoomies) > 1 {
			combat.GymLoomies = combat.GymLoomies[1:]
		} else {
			combat.GymLoomies = make([]interfaces.CombatLoomie, 0)
		}

		// Notify the user that the gym loomie was weakened
		combat.SendMessage(WsMessage{
			Type:    "GYM_LOOMIE_WEAKENED",
			Message: fmt.Sprintf("Enemy loomie %s was weakened", gymLoomie.Name),
			Payload: map[string]interface{}{
				"loomie_id": wenakenedLoomieId,
			},
		})

		// Check if the player won the battle
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

		// Notify the user that the current gym loomie was changed
		combat.SendMessage(WsMessage{
			Type:    "UPDATE_GYM_LOOMIE",
			Message: fmt.Sprintf("Enemy loomie %s is now the gym's active loomie", combat.CurrentGymLoomie.Name),
			Payload: map[string]interface{}{
				"loomie": combat.CurrentGymLoomie,
			},
		})

		// Add experience to the player loomies that fought the gym loomie
		handleGymLoomieWeakened(combat, wenakenedLoomieId)
		return
	} else {
		// Notify the user that the gym loomie hp was updated
		combat.SendMessage(WsMessage{
			Type:    "UPDATE_GYM_LOOMIE_HP",
			Message: fmt.Sprintf("Enemy loomie %s received %d damage", gymLoomie.Name, calculatedAttack),
			Payload: map[string]interface{}{
				"loomie_id": gymLoomie.Id,
				"hp":        gymLoomie.BoostedHp,
			},
		})
	}
}

// handleGymLoomieWeakened handles the "event" when a gym loomie is weakened by the player to add experience to the player loomies that fought the gym loomie
func handleGymLoomieWeakened(combat *WsCombat, weakenedLoomieId primitive.ObjectID) {
	// TODO: Silvia, you should add the functionality to add experience to the player
	// loomies locally and also in the database.

	fmt.Println("Handling Gym Loomie Weakened Event for:", weakenedLoomieId)
	foughtWith := combat.FoughtGymLoomies[weakenedLoomieId]

	for _, playerLoomieId := range foughtWith {
		// TODO: Add experience to the player loomie
		fmt.Println("Adding experience to player loomie:", playerLoomieId)
	}

	fmt.Println("Weakened Event Handled")
}

// handleClearDodgeChannel Clears the dodge channel to avoid collisions between attacks
func handleClearDodgeChannel(combat *WsCombat) {
	for len(combat.Dodges) > 0 {
		<-combat.Dodges
	}
}
