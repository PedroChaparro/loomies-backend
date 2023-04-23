package combat

import (
	"fmt"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
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

	// Get the current loomies after the timeout to prevent desync
	gymLoomie := combat.CurrentGymLoomie
	playerLoomie := combat.CurrentPlayerLoomie

	// Send the dodge message to the client if the attack was dodged
	if wasAttackDodged {
		combat.SendMessage(WsMessage{
			Type:    "GYM_ATTACK_DODGED",
			Message: fmt.Sprintf("Your loomie %s dodged the attack", playerLoomie.Name),
		})

		return
	}

	// Check if the types of the loomie were obtained before
	cacheTypeStrongAgainst(gymLoomie.Types, combat)

	// Calculate the damage
	calculatedAttack := calculateAttack(gymLoomie, playerLoomie)

	// Reduce the player loomie hp
	playerLoomie.BoostedHp -= calculatedAttack

	// Check if the player loomie was weakened
	if playerLoomie.BoostedHp <= 0 {
		weakenedLoomieId := playerLoomie.Id

		// Reduce the alive player loomies count
		combat.AlivePlayerLoomies--

		// Notify the user that the loomie was weakened
		combat.SendMessage(WsMessage{
			Type:    "USER_LOOMIE_WEAKENED",
			Message: fmt.Sprintf("Your loomie %s was weakened", playerLoomie.Name),
			Payload: map[string]interface{}{
				"loomie_id": weakenedLoomieId,
			},
		})

		// Check if the player loose the battle
		if combat.AlivePlayerLoomies == 0 {
			combat.SendMessage(WsMessage{
				Type:    "USER_HAS_LOST",
				Message: "You have lost the battle. Try fusioning your loomies or caught more loomies to improve your team",
			})

			combat.Close <- true
			return
		}

		for index := range combat.PlayerLoomies {
			if combat.PlayerLoomies[index].BoostedHp > 0 {
				// Update the current gym loomie
				combat.CurrentPlayerLoomie = &combat.PlayerLoomies[index]
				break
			}
		}

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
		// Reduce the alive gym loomies count
		weakenedLoomie := gymLoomie
		combat.AliveGymLoomies--

		// Notify the user that the gym loomie was weakened
		combat.SendMessage(WsMessage{
			Type:    "GYM_LOOMIE_WEAKENED",
			Message: fmt.Sprintf("Enemy loomie %s was weakened", gymLoomie.Name),
			Payload: map[string]interface{}{
				"loomie_id": weakenedLoomie.Id,
			},
		})

		// Check if the player won the battle
		if combat.AliveGymLoomies == 0 {
			combat.SendMessage(WsMessage{
				Type:    "USER_HAS_WON",
				Message: "You have won the battle. Now you own this gym",
			})

			handlePlayerVictory(combat)
			return
		}

		for index := range combat.GymLoomies {
			if combat.GymLoomies[index].BoostedHp > 0 {
				// Update the current gym loomie
				combat.CurrentGymLoomie = &combat.GymLoomies[index]
				break
			}
		}

		// Notify the user that the current gym loomie was changed
		combat.SendMessage(WsMessage{
			Type:    "UPDATE_GYM_LOOMIE",
			Message: fmt.Sprintf("Enemy loomie %s is now the gym's active loomie", combat.CurrentGymLoomie.Name),
			Payload: map[string]interface{}{
				"loomie": combat.CurrentGymLoomie,
			},
		})

		// Add experience to the player loomies that fought the gym loomie
		handleGymLoomieWeakened(combat, weakenedLoomie.Id, weakenedLoomie.Level)
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
func handleGymLoomieWeakened(combat *WsCombat, weakenedLoomieId primitive.ObjectID, levelWeakenedLoomieId int) {
	// Obtains Loomies who weakened the enemy Loomie
	foughtWith := combat.FoughtGymLoomies[weakenedLoomieId]

	// Calculates exp of Loomie weakened and its third part. It is divided in # of Loomies
	expWeakenedLoomieId := utils.GetRequiredExperience(levelWeakenedLoomieId)
	experienceToSet := (expWeakenedLoomieId / 3) / float64(len(foughtWith))

	// adds the experience to each Loomie in foughtWith
	for index := range foughtWith {
		playerLoomiePointer := foughtWith[index]

		// usefull if the loomie level up
		preLevel := playerLoomiePointer.Level

		// calculates and sets new exp and lvl locally
		playerLoomiePointer.Experience, playerLoomiePointer.Level = calculateLevelAndExperience(playerLoomiePointer.Experience, experienceToSet, playerLoomiePointer.Level)

		// updates and sets new exp and lvl in db
		models.UpdateLoomiesExpAndLvl(combat.PlayerID, playerLoomiePointer)

		combat.SendMessage(WsMessage{
			Type:    "UPDATE_USER_LOOMIE_EXP",
			Message: fmt.Sprintf("Loomie %s received %.4f of experience", playerLoomiePointer.Name, experienceToSet),
			Payload: map[string]interface{}{
				"loomie": playerLoomiePointer.Id,
			},
		})

		// if loomie level up, there is an update in its stats
		if playerLoomiePointer.Level-preLevel != 0 {
			updateStatsDuringWeakenedEvent(playerLoomiePointer)
			combat.SendMessage(WsMessage{
				Type:    "UPDATE_PLAYER_LOOMIE",
				Message: fmt.Sprintf("Loomie %s received an update of hp, attack and defense", playerLoomiePointer.Name),
				Payload: map[string]interface{}{
					"loomie": playerLoomiePointer,
				},
			})
		}
	}
}

// handlePlayerVictory handles the "event" when the player wins the battle
func handlePlayerVictory(combat *WsCombat) {
	gymId, _ := primitive.ObjectIDFromHex(combat.GymID)
	gymInfo, err := models.GetPopulatedGymFromId(gymId, combat.PlayerID)

	if err != nil {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[INTERNAL SERVER ERROR] Error obtaining gym info",
		})

		return
	}

	// Updates the loomie team of the new owner with an empty array
	err = models.ReplaceLoomieTeam(combat.PlayerID, []primitive.ObjectID{})
	if err != nil {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[INTERNAL SERVER ERROR] Error updating Loomie Team of user",
		})
	}

	// Obtains ids of new protectors
	newGymProtectors := []primitive.ObjectID{}
	currentGymProtectors := []primitive.ObjectID{}

	for _, playerLoomie := range combat.PlayerLoomies {
		newGymProtectors = append(newGymProtectors, playerLoomie.Id)
	}

	for _, gymLoomie := range combat.GymLoomies {
		currentGymProtectors = append(currentGymProtectors, gymLoomie.Id)
	}

	if gymInfo.Owner != "" {
		// Updates the gym old protectors
		err = models.UpdateLoomiesBusyState(currentGymProtectors, false)
		if err != nil {
			combat.SendMessage(WsMessage{
				Type:    "ERROR",
				Message: "[INTERNAL SERVER ERROR] Error updating current gym protectors status",
			})
		}
	} else {
		// Removes the gym old protectors
		err = models.RemoveLoomieTeam(currentGymProtectors)
		if err != nil {
			combat.SendMessage(WsMessage{
				Type:    "ERROR",
				Message: "[INTERNAL SERVER ERROR] Error deleting current gym protectors",
			})
		}
	}

	// Updates the gym news protectors and owner
	err = models.UpdateGymProtectorsAndOwner(gymId, newGymProtectors, combat.PlayerID)
	if err != nil {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[INTERNAL SERVER ERROR] Error updating Loomie Team of user and its owner",
		})
	}

	// Updates the gym news protectors, is_busy propierties
	err = models.UpdateLoomiesBusyState(newGymProtectors, true)
	if err != nil {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[INTERNAL SERVER ERROR] Error updating Loomie Team is_busy field",
		})
	}

	combat.Close <- true
}

// handleUseItem handles the use of an item by the player
func handleUseItem(combat *WsCombat, message WsMessage) {
	// Get the item id from the messge payload
	payload := message.Payload
	itemId := fmt.Sprint(payload["item_id"])

	// Check the item is not null
	if itemId == "" {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[BAD REQUEST] Item id is required",
		})

		return
	}

	// Check if the item is a valid mongo id
	itemMongoId, err := primitive.ObjectIDFromHex(itemId)

	if err != nil {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[BAD REQUEST] Item id is not valid",
		})

		return
	}

	// Check the item exists in the user inventory
	item, err := models.GetItemFromUserInventory(combat.PlayerID, itemMongoId, false)

	if err != nil {
		if err.Error() == "USER_DOES_NOT_OWN_ITEM" {
			combat.SendMessage(WsMessage{
				Type:    "ERROR",
				Message: "[BAD REQUEST] You don't own this item",
			})

			return
		}

		if err.Error() == "ITEM_NOT_FOUND" {
			combat.SendMessage(WsMessage{
				Type:    "ERROR",
				Message: "[BAD REQUEST] Item does not exist or is not a combat item",
			})

			return
		}

		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[INTERNAL SERVER ERROR] Error getting the item from the user inventory",
		})

		return
	}

	// Apply the item
	err = applyItem(combat.PlayerID, &item, combat.CurrentPlayerLoomie)

	if err != nil {
		// If the loomie does not need healing, send a message to the user
		if err.Error() == "HEALING_NOT_NEEDED" {
			combat.SendMessage(WsMessage{
				Type:    "ERROR",
				Message: "[BAD REQUEST] The loomie is not damaged",
			})

			return
		}

		if err.Error() == "SERVER_ERROR" {
			combat.SendMessage(WsMessage{
				Type:    "ERROR",
				Message: "[INTERNAL SERVER ERROR] There was an error using the item. Please try again later",
			})

			return
		}

		// If the item is not supported, send a message to the user
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[BAD REQUEST] The item is not supported",
		})

		return
	}

	// Decrement the item from the user inventory
	err = models.DecrementItemFromUserInventory(combat.PlayerID, itemMongoId, 1)

	if err != nil {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[INTERNAL SERVER ERROR] There was an error using the item. Please try again later",
		})

		return
	}

	// Send the message to the user
	combat.SendMessage(WsMessage{
		Type:    "UPDATE_PLAYER_LOOMIE",
		Message: fmt.Sprintf("Loomie: %s received the item: %s", combat.CurrentPlayerLoomie.Name, item.Name),
		Payload: map[string]interface{}{
			"loomie": combat.CurrentPlayerLoomie,
		},
	})
}

// handleChangeLoomie handles the change of the player loomie
func handleChangeLoomie(combat *WsCombat, message WsMessage) {
	// Get the loomie id from the message payload
	payload := message.Payload
	loomieId := fmt.Sprint(payload["loomie_id"])

	// Check the loomie id is not null
	if loomieId == "" {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[BAD REQUEST] Loomie id is required",
		})

		return
	}

	// Check if the loomie id is a valid mongo id
	loomieMongoId, err := primitive.ObjectIDFromHex(loomieId)

	if err != nil {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[BAD REQUEST] Loomie id is not valid",
		})

		return
	}

	// Check the loomie is in the player loomies and is not weakened
	var loomieExists, loomieAlive bool
	var loomieIndex int

	for index := range combat.PlayerLoomies {
		if combat.PlayerLoomies[index].Id == loomieMongoId {
			loomieExists = true

			if combat.PlayerLoomies[index].BoostedHp > 0 {
				loomieAlive = true
			}

			loomieIndex = index
			break
		}
	}

	if !loomieExists {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[BAD REQUEST] Loomie does not exist",
		})

		return
	}

	if !loomieAlive {
		combat.SendMessage(WsMessage{
			Type:    "ERROR",
			Message: "[BAD REQUEST] Loomie is weakened",
		})

		return
	}

	// Change the current player loomie
	combat.CurrentPlayerLoomie = &combat.PlayerLoomies[loomieIndex]

	// Send the message to the user
	combat.SendMessage(WsMessage{
		Type:    "UPDATE_PLAYER_LOOMIE",
		Message: fmt.Sprintf("Loomie: %s is now the current player loomie", combat.CurrentPlayerLoomie.Name),
		Payload: map[string]interface{}{
			"loomie": combat.CurrentPlayerLoomie,
		},
	})
}

// handleClearDodgeChannel Clears the dodge channel to avoid collisions between attacks
func handleClearDodgeChannel(combat *WsCombat) {
	for len(combat.Dodges) > 0 {
		<-combat.Dodges
	}
}
