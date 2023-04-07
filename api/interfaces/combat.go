package interfaces

import (
	"math"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ------------------------------------------
// In this file we define the interfaces for the combat
// and the methods for the CombatLoomie to boost the stats
// ------------------------------------------
type CombatLoomie struct {
	Id             primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Serial         int                `json:"serial"      bson:"serial"`
	Name           string             `json:"name"      bson:"name"`
	Types          []string           `json:"types"     bson:"types"`
	Rarity         string             `json:"rarity"     bson:"rarity"`
	BaseHp         int                `json:"hp"     bson:"hp"`
	BaseAttack     int                `json:"attack"     bson:"attack"`
	BaseDefense    int                `json:"defense"     bson:"defense"`
	BoostedHp      int                `json:"boosted_hp"     bson:"boosted_hp"`
	BoostedAttack  int                `json:"boosted_attack"     bson:"boosted_attack"`
	BoostedDefense int                `json:"boosted_defense"     bson:"boosted_defense"`
	Level          int                `json:"level"     bson:"level"`
	Experience     float64            `json:"experience"     bson:"experience"`
	IsBusy         bool               `json:"is_busy"     bson:"is_busy"`
}

// ToCombatLoomie Converts a user loomie to a combat loomie boosting the stats according to the level
func (normalCaughtLoomie *UserLoomiesRes) ToCombatLoomie() *CombatLoomie {
	experienceFactor := 1.0 + ((1.0 / 8.0) * (float64(normalCaughtLoomie.Level) - 1.0))

	return &CombatLoomie{
		Id:             normalCaughtLoomie.Id,
		Serial:         normalCaughtLoomie.Serial,
		Name:           normalCaughtLoomie.Name,
		Types:          normalCaughtLoomie.Types,
		Rarity:         normalCaughtLoomie.Rarity,
		BaseHp:         normalCaughtLoomie.Hp,
		BaseAttack:     normalCaughtLoomie.Attack,
		BaseDefense:    normalCaughtLoomie.Defense,
		BoostedHp:      int(math.Floor(float64(normalCaughtLoomie.Hp) * experienceFactor)),
		BoostedAttack:  int(math.Floor(float64(normalCaughtLoomie.Attack) * experienceFactor)),
		BoostedDefense: int(math.Floor(float64(normalCaughtLoomie.Defense) * experienceFactor)),
		Level:          normalCaughtLoomie.Level,
		Experience:     normalCaughtLoomie.Experience,
		IsBusy:         normalCaughtLoomie.IsBusy,
	}
}

// ApplyPainKillers Boosts the hp of the loomie by 50
func (loomie *CombatLoomie) ApplyPainKillers() {
	loomie.BoostedHp += 50
}

// ApplySmallAidKit Boosts the hp of the loomie by 100
func (loomie *CombatLoomie) ApplySmallAidKit() {
	loomie.BoostedHp += 100
}

// ApplyBigAidKit Boosts the hp of the loomie by restoring it to the base hp according to the level
func (loomie *CombatLoomie) ApplyBigAidKit() {
	experienceFactor := 1.0 + ((1.0 / 8.0) * (float64(loomie.Level) - 1.0))
	loomie.BoostedHp = int(math.Floor(float64(loomie.BaseHp) * experienceFactor))
}

// ApplyDefibrillator Revives the loomie by setting the hp to half of the base hp according to the level
func (loomie *CombatLoomie) ApplyDefibrillator() {
	experienceFactor := 1.0 + ((1.0 / 8.0) * (float64(loomie.Level) - 1.0))
	loomie.BoostedHp = int(math.Floor(float64(loomie.BaseHp)*experienceFactor)) / 2
}

// ApplySteroidsInjection Boosts the attack of the loomie by 20% of the base attack according to the level
func (loomie *CombatLoomie) ApplySteroidsInjection() {
	experienceFactor := 1.0 + ((1.0 / 8.0) * (float64(loomie.Level) - 1.0))
	loomie.BoostedAttack += int(math.Floor(float64(loomie.BaseAttack)*experienceFactor)) / 5
}

// ApplyVitamins Boosts hp of the loomie by 20% of the base hp according to the level
func (loomie *CombatLoomie) ApplyVitamins() {
	experienceFactor := 1.0 + ((1.0 / 8.0) * (float64(loomie.Level) - 1.0))
	loomie.BoostedHp += int(math.Floor(float64(loomie.BaseHp)*experienceFactor)) / 5
}

// ApplyUnknownBevarage Increases the level of the loomie by 1 and updates the loomie's stats
func (loomie *CombatLoomie) ApplyUnknownBevarage() {
	// Keep the previous boosts
	experienceFactor := 1.0 + ((1.0 / 8.0) * (float64(loomie.Level) - 1.0))
	previousInitialBoostedHp := int(math.Floor(float64(loomie.BaseHp) * experienceFactor))
	previousHpBoosts := loomie.BoostedHp - previousInitialBoostedHp

	previousInitialBoostedAttack := int(math.Floor(float64(loomie.BaseAttack) * experienceFactor))
	previousAttackBoosts := loomie.BoostedAttack - previousInitialBoostedAttack

	previousInitialBoostedDefense := int(math.Floor(float64(loomie.BaseDefense) * experienceFactor))
	previousDefenseBoosts := loomie.BoostedDefense - previousInitialBoostedDefense

	// Increment the level and update the stats
	loomie.Level++
	experienceFactor = 1.0 + ((1.0 / 8.0) * (float64(loomie.Level) - 1.0))
	loomie.BoostedHp = int(math.Floor(float64(loomie.BaseHp)*experienceFactor)) + previousHpBoosts
	loomie.BoostedAttack = int(math.Floor(float64(loomie.BaseAttack)*experienceFactor)) + previousAttackBoosts
	loomie.BoostedDefense = int(math.Floor(float64(loomie.BaseDefense)*experienceFactor)) + previousDefenseBoosts
}
