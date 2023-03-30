package configuration

import (
	"github.com/PedroChaparro/loomies-backend/combat"
	"go.mongodb.org/mongo-driver/mongo"
)

type TGlobals struct {
	Loaded                      bool
	MongoClient                 *mongo.Client
	AccessTokenSecret           string
	RefreshTokenSecret          string
	WsTokenSecret               string
	WsHub                       *combat.WsHub
	MinLoomiesGenerationTimeout int
	MaxLoomiesGenerationTimeout int
	MinLoomiesGenerationAmount  int
	MaxLoomiesGenerationAmount  int
	LoomiesGenerationRadius     float64
	MaxLoomiesPerZone           int
	// Global settings to calculate the experience required to level up
	MinLoomieRequiredExperience float64
	LoomieExperienceFactor      float64
}
