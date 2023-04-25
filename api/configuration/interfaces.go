package configuration

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type TGlobals struct {
	Environment                 string
	Loaded                      bool
	MongoClient                 *mongo.Client
	AccessTokenSecret           string
	RefreshTokenSecret          string
	WsTokenSecret               string
	WildLoomiesTTL              int
	MinLoomiesGenerationTimeout int
	MaxLoomiesGenerationTimeout int
	MinLoomiesGenerationAmount  int
	MaxLoomiesGenerationAmount  int
	LoomiesGenerationRadius     float64
	MaxLoomiesPerZone           int
	// Global settings to calculate the experience required to level up
	MinLoomieRequiredExperience float64
	LoomieExperienceFactor      float64
	// Global settings to be used on combats
	MinCombatAttackTimeout int
	MaxCombatAttackTimeout int
	CombatChallengeTimeout int
}
