package combat

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
)

// getRandomInt returns a random integer between min and max (both included)
// Note: This function is duplicated in api/utils/helpers.go because of the
// cyclic dependency
func getRandomInt(min int, max int) int {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return r.Intn(max-min) + min
}

// isTypeStrongAgainst returns true if the atacking type is strong against the defending type
func isTypeStrongAgainst(atackingType string, defendingTypes []string) bool {
	for _, strongAgainst := range GlobalWsHub.CachedStrongAgainst[atackingType] {
		for _, defendingType := range defendingTypes {
			if strongAgainst == defendingType {
				return true
			}
		}
	}

	return false
}

// cacheTypeStrongAgainst caches the strong against types if they are not cached yet
func cacheTypeStrongAgainst(loomieTypes []string, combat *WsCombat) {
	// For each type
	for _, value := range loomieTypes {
		// Check if the type was cached before
		_, cached := GlobalWsHub.CachedStrongAgainst[value]

		// If the type was not obtained before, get it from the database and cache it
		if !cached {
			typeDetails, err := models.GetLoomieTypeDetailsByName(value)

			if err != nil {
				combat.SendMessage(WsMessage{
					Type:    "ERROR",
					Message: "[INTERNAL SERVER ERROR] Error getting the loomie type details",
				})

				return
			}

			GlobalWsHub.CachedStrongAgainst[value] = make([]string, 0)
			GlobalWsHub.CachedStrongAgainst[value] = typeDetails.StrongAgainst
		}
	}
}

// calculateAttack calculates the final attack of the atacking loomie
func calculateAttack(atackingLoomie, defendingLoomie *interfaces.CombatLoomie) int {
	// Initial attack value
	finalAttack := atackingLoomie.BoostedAttack
	minAttack := float64(atackingLoomie.BoostedAttack) * 0.1

	// Increment the attack if the loomie is strong against the defending loomie
	for _, atackingLoomieType := range atackingLoomie.Types {
		if isTypeStrongAgainst(atackingLoomieType, defendingLoomie.Types) {
			finalAttack *= 2
			break
		}
	}

	// Apply the user loomie defense
	finalAttack -= finalAttack * (defendingLoomie.BoostedDefense / 100)
	finalAttack = int(math.Max(float64(finalAttack), minAttack))
	return finalAttack
}

// applyItem Applies the item to the loomie by its serial
func applyItem(item *interfaces.PopulatedInventoryItem, loomie *interfaces.CombatLoomie) error {
	switch item.Serial {
	// Painkiller
	case 1:
		loomie.ApplyPainKillers()
	// Small aid kit
	case 2:
		loomie.ApplySmallAidKit()
	// Big aid kit
	case 3:
		loomie.ApplyBigAidKit()
		// Defibrillator
	case 4:
		loomie.ApplyDefibrillator()
		// Steroids injection
	case 5:
		loomie.ApplySteroidsInjection()
		// Vitamins
	case 6:
		loomie.ApplyVitamins()
		// Unknown bevarage
	case 7:
		loomie.ApplyUnknownBevarage()
	default:
		return fmt.Errorf("NON_SUPPORTED_ITEM")
	}

	return nil
}
