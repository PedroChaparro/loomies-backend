package combat

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func applyItem(userId primitive.ObjectID, item *interfaces.PopulatedInventoryItem, loomie *interfaces.CombatLoomie) error {
	switch item.Serial {
	// Painkiller
	case 1:
		wasApplied := loomie.ApplyPainKillers()
		if !wasApplied {
			return fmt.Errorf("HEALING_NOT_NEEDED")
		}
	// Small aid kit
	case 2:
		wasApplied := loomie.ApplySmallAidKit()
		if !wasApplied {
			return fmt.Errorf("HEALING_NOT_NEEDED")
		}
	// Big aid kit
	case 3:
		wasApplied := loomie.ApplyBigAidKit()
		if !wasApplied {
			return fmt.Errorf("HEALING_NOT_NEEDED")
		}
		// Defibrillator
	case 4:
		wasApplied := loomie.ApplyDefibrillator()
		if !wasApplied {
			return fmt.Errorf("HEALING_NOT_NEEDED")
		}
		// Steroids injection
	case 5:
		loomie.ApplySteroidsInjection()
		// Vitamins
	case 6:
		loomie.ApplyVitamins()
		// Unknown bevarage
	case 7:
		loomie.ApplyUnknownBevarage()
		err := models.IncrementLoomieLevel(userId, loomie.Id, 1)

		if err != nil {
			return fmt.Errorf("SERVER_ERROR")
		}
	default:
		return fmt.Errorf("NON_SUPPORTED_ITEM")
	}

	return nil
}

// TODO generalize?
// calculateLevelAndExperience calculates what is lvl and experience of a Loomie that weakened another one
func calculateLevelAndExperience(loomieExperience float64, availableExperience float64, loomieLevel int) (float64, int) {
	var experienceToAdd, neededExperienceToNextLevel float64

	// Check if the loomie has leveled up
	for (loomieExperience + availableExperience) >= utils.GetRequiredExperience(loomieLevel+1) {
		neededExperienceToNextLevel = utils.GetRequiredExperience(loomieLevel+1) - loomieExperience
		experienceToAdd = math.Min(availableExperience, neededExperienceToNextLevel)
		experienceToAdd = utils.FixeFloat(experienceToAdd, 4)
		loomieLevel++
		loomieExperience = 0
		availableExperience -= experienceToAdd
	}

	loomieExperience += availableExperience

	resultExp := utils.FixeFloat(loomieExperience, 4)
	return resultExp, loomieLevel
}

// UpdateStatsDuringWeakenedEvent updates maxhp, hp, attack and defense if a loomie advance in lvl during a weakened event
func updateStatsDuringWeakenedEvent(loomieToUpdate *interfaces.CombatLoomie) {
	// Tracking previous boosts in MaxHP, Attack, Defense
	initialMaxHp := calulateExperienceFactorStats(loomieToUpdate.BaseHp, loomieToUpdate.Level-1)
	previousMaxHpBoosts := loomieToUpdate.MaxHp - initialMaxHp

	initialBoostedAttack := calulateExperienceFactorStats(loomieToUpdate.BaseAttack, loomieToUpdate.Level-1)
	previousAttackBoosts := loomieToUpdate.BoostedAttack - initialBoostedAttack

	initialBoostedDefense := calulateExperienceFactorStats(loomieToUpdate.BaseDefense, loomieToUpdate.Level-1)
	previousDefenseBoosts := loomieToUpdate.BoostedDefense - initialBoostedDefense

	// Update previous boosts in MaxHP, Attack, Defense
	loomieToUpdate.MaxHp = calulateExperienceFactorStats(loomieToUpdate.BaseHp, loomieToUpdate.Level) + previousMaxHpBoosts
	loomieToUpdate.BoostedAttack = calulateExperienceFactorStats(loomieToUpdate.BaseAttack, loomieToUpdate.Level) + previousAttackBoosts
	loomieToUpdate.BaseDefense = calulateExperienceFactorStats(loomieToUpdate.BaseDefense, loomieToUpdate.Level) + previousDefenseBoosts

	// Here the new level boost (+1/8 of the base hp) is incremented
	possibleHpIncrement := int(math.Floor(float64(loomieToUpdate.BaseHp)) * (1.0 / 8.0))
	possibleBoostedHp := float64(loomieToUpdate.BoostedHp) + float64(possibleHpIncrement)
	loomieToUpdate.BoostedHp = int(math.Min(possibleBoostedHp, float64(loomieToUpdate.MaxHp)))
}

// CalulateExperienceFactorStats helper of updateStatsDuringWeakenedEvent
func calulateExperienceFactorStats(baseStat int, level int) int {
	experienceFactor := (1.0 + ((1.0 / 8.0) * (float64(level) - 1.0)))
	return int(math.Floor(float64(baseStat) * experienceFactor))
}
