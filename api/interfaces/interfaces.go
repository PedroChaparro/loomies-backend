package interfaces

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Globals struct {
	Loaded                      bool
	MongoClient                 *mongo.Client
	AccessTokenSecret           string
	RefreshTokenSecret          string
	MinLoomiesGenerationTimeout int
	MaxLoomiesGenerationTimeout int
	MinLoomiesGenerationAmount  int
	MaxLoomiesGenerationAmount  int
	LoomiesGenerationRadius     float64
	MaxLoomiesPerZone           int
}

type Coordinates struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

type Zone struct {
	Id             primitive.ObjectID   `json:"_id" bson:"_id"`
	LeftFrontier   float64              `json:"leftFrontier" bson:"leftFrontier"`
	RightFrontier  float64              `json:"rightFrontier" bson:"rightFrontier"`
	TopFrontier    float64              `json:"topFrontier" bson:"topFrontier"`
	BottomFrontier float64              `json:"bottomFrontier" bson:"bottomFrontier"`
	Number         int                  `json:"number" bson:"number"`
	Coordinates    string               `json:"coordinates" bson:"coordinates"`
	Gym            primitive.ObjectID   `json:"gym" bson:"gym"`
	Loomies        []primitive.ObjectID `json:"loomies" bson:"loomies"`
}

type ZoneWithGyms struct {
	Id             primitive.ObjectID   `json:"_id" bson:"_id"`
	LeftFrontier   float64              `json:"leftFrontier" bson:"leftFrontier"`
	RightFrontier  float64              `json:"rightFrontier" bson:"rightFrontier"`
	TopFrontier    float64              `json:"topFrontier" bson:"topFrontier"`
	BottomFrontier float64              `json:"bottomFrontier" bson:"bottomFrontier"`
	Number         int                  `json:"number" bson:"number"`
	Coordinates    string               `json:"coordinates" bson:"coordinates"`
	Gym            primitive.ObjectID   `json:"gym" bson:"gym"`
	Gyms           []Gym                `json:"gyms" bson:"gyms"`
	Loomies        []primitive.ObjectID `json:"loomies" bson:"loomies"`
}

type ZoneWithPopulatedLoomies struct {
	Id               primitive.ObjectID   `json:"_id" bson:"_id"`
	LeftFrontier     float64              `json:"leftFrontier" bson:"leftFrontier"`
	RightFrontier    float64              `json:"rightFrontier" bson:"rightFrontier"`
	TopFrontier      float64              `json:"topFrontier" bson:"topFrontier"`
	BottomFrontier   float64              `json:"bottomFrontier" bson:"bottomFrontier"`
	Number           int                  `json:"number" bson:"number"`
	Coordinates      string               `json:"coordinates" bson:"coordinates"`
	Gym              primitive.ObjectID   `json:"gym" bson:"gym"`
	Loomies          []primitive.ObjectID `json:"loomies" bson:"loomies"`
	PopulatedLoomies []WildLoomie         `json:"populated_loomies" bson:"populated_loomies"`
}

type GymRewardItem struct {
	RewardCollection string             `json:"reward_collection" bson:"reward_collection"`
	RewardId         primitive.ObjectID `json:"reward_id" bson:"reward_id"`
	RewardQuantity   int                `json:"reward_quantity" bson:"reward_quantity"`
}

type Gym struct {
	Id                    primitive.ObjectID   `json:"_id" bson:"_id"`
	Latitude              float64              `json:"latitude"      bson:"latitude"`
	Longitude             float64              `json:"longitude"      bson:"longitude"`
	Name                  string               `json:"name"      bson:"name"`
	Owner                 primitive.ObjectID   `json:"owner,omitempty"      bson:"owner,omitempty"`
	CurrentPlayersRewards []GymRewardItem      `json:"current_players_rewards"      bson:"current_players_rewards"`
	CurrentOwnerRewards   []GymRewardItem      `json:"current_owner_rewards"      bson:"current_owner_rewards"`
	RewardsClaimedBy      []primitive.ObjectID `json:"rewards_claimed_by"      bson:"rewards_claimed_by"`
}

type Item struct {
	Id                    primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Name                  string             `json:"name"      bson:"name"`
	Description           string             `json:"description"      bson:"description"`
	Target                string             `json:"target"      bson:"target"`
	IsCombatItem          bool               `json:"is_combat_item"      bson:"is_combat_item"`
	GymRewardChancePlayer float64            `json:"gym_reward_chance_player"      bson:"gym_reward_chance_player"`
	GymRewardChanceOwner  float64            `json:"gym_reward_chance_owner"      bson:"gym_reward_chance_owner"`
	MinRewardQuantity     int                `json:"min_reward_quantity"      bson:"min_reward_quantity"`
	MaxRewardQuantity     int                `json:"max_reward_quantity"      bson:"max_reward_quantity"`
}

type Loomball struct {
	Id                    primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Name                  string             `json:"name"      bson:"name"`
	EffectiveUntil        int64              `json:"effective_until"      bson:"effective_until"`
	DecayUntil            int64              `json:"decay_until"      bson:"decay_until"`
	MinimumProbability    float64            `json:"minimum_probability"      bson:"minimum_probability"`
	GymRewardChancePlayer float64            `json:"gym_reward_chance_player"      bson:"gym_reward_chance_player"`
	GymRewardChanceOwner  float64            `json:"gym_reward_chance_owner"      bson:"gym_reward_chance_owner"`
	MinRewardQuantity     int                `json:"min_reward_quantity"      bson:"min_reward_quantity"`
	MaxRewardQuantity     int                `json:"max_reward_quantity"      bson:"max_reward_quantity"`
}

type InventoryItem struct {
	ItemCollection string             `json:"item_collection" bson:"item_collection"`
	ItemId         primitive.ObjectID `json:"item_id" bson:"item_id"`
	ItemQuantity   int                `json:"item_quantity" bson:"item_quantity"`
}

type User struct {
	Id                              primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Username                        string               `json:"username"      bson:"username"`
	Email                           string               `json:"email"     bson:"email"`
	Password                        string               `json:"password"  bson:"password"`
	Items                           []InventoryItem      `json:"items"     bson:"items"`
	Loomies                         []primitive.ObjectID `json:"loomies"   bson:"loomies"`
	IsVerified                      bool                 `json:"isVerified"   bson:"isVerified"`
	CurrentLoomiesGenerationTimeout int64                `json:"currentLoomiesGenerationTimeout"   bson:"currentLoomiesGenerationTimeout"`
	LastLoomieGenerationTime        int64                `json:"lastLoomieGenerationTime"   bson:"lastLoomieGenerationTime"`
}

type LoomieRarity struct {
	Id          primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Name        string             `json:"name"      bson:"name"`
	SpawnChance float64            `json:"spawn_chance"      bson:"spawn_chance"`
}

type LoomieType struct {
	Id            primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Name          string               `json:"name"      bson:"name"`
	StrongAgainst []primitive.ObjectID `json:"strong_against"      bson:"strong_against"`
}

// BaseLoomie is the "template" for a loomie
type BaseLoomies struct {
	Id          primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Serial      int                  `json:"serial"      bson:"serial"`
	Name        string               `json:"name"      bson:"name"`
	Types       []primitive.ObjectID `json:"types"     bson:"types"`
	Rarity      primitive.ObjectID   `json:"rarity"     bson:"rarity"`
	BaseHp      int                  `json:"base_hp"     bson:"base_hp"`
	BaseAttack  int                  `json:"base_attack"     bson:"base_attack"`
	BaseDefense int                  `json:"base_defense"     bson:"base_defense"`
}

type BaseLoomiesWithPopulatedRarity struct {
	Id              primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Serial          int                  `json:"serial"      bson:"serial"`
	Name            string               `json:"name"      bson:"name"`
	Types           []primitive.ObjectID `json:"types"     bson:"types"`
	BaseHp          int                  `json:"base_hp"     bson:"base_hp"`
	BaseAttack      int                  `json:"base_attack"     bson:"base_attack"`
	BaseDefense     int                  `json:"base_defense"     bson:"base_defense"`
	Rarity          primitive.ObjectID   `json:"rarity"     bson:"rarity"`
	PopulatedRarity LoomieRarity         `json:"populated_rarity"     bson:"populated_rarity"`
}

// WildLoomie is a global loomie that can be found in the wild to be caught by any user
type WildLoomie struct {
	Id          primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Serial      int                  `json:"serial"      bson:"serial"`
	Name        string               `json:"name"      bson:"name"`
	Types       []primitive.ObjectID `json:"types"     bson:"types"`
	Rarity      primitive.ObjectID   `json:"rarity"     bson:"rarity"`
	HP          int                  `json:"hp"     bson:"hp"`
	Attack      int                  `json:"attack"     bson:"attack"`
	Defense     int                  `json:"defense"     bson:"defense"`
	ZoneId      primitive.ObjectID   `json:"zone_id"     bson:"zone_id"`
	Latitude    float64              `json:"latitude"     bson:"latitude"`
	Longitude   float64              `json:"longitude"     bson:"longitude"`
	GeneratedAt int64                `json:"generated_at"     bson:"generated_at"`
}
