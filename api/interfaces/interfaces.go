package interfaces

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	Protectors            []primitive.ObjectID `json:"protectors"      bson:"protectors"`
	CurrentPlayersRewards []GymRewardItem      `json:"current_players_rewards"      bson:"current_players_rewards"`
	CurrentOwnerRewards   []GymRewardItem      `json:"current_owners_rewards"      bson:"current_owners_rewards"`
	RewardsClaimedBy      []primitive.ObjectID `json:"rewards_claimed_by"      bson:"rewards_claimed_by"`
}

// Struct to keep only the necessary data from the Caught Loomies collection
type GymProtector struct {
	Serial int    `json:"serial" bson:"serial"`
	Name   string `json:"name" bson:"name"`
	Level  int    `json:"level" bson:"level"`
}

// Auxiliar struct to parse the database response
type PopulatedGymAux struct {
	Id               primitive.ObjectID   `json:"_id" bson:"_id"`
	Name             string               `json:"name"      bson:"name"`
	Owner            []User               `json:"owner,omitempty"      bson:"owner,omitempty"`
	Protectors       []CaughtLoomie       `json:"protectors"      bson:"protectors"`
	RewardsClaimedBy []primitive.ObjectID `json:"rewards_claimed_by"      bson:"rewards_claimed_by"`
}

// Final struct to be returned to the client
type PopulatedGym struct {
	Id               primitive.ObjectID `json:"_id" bson:"_id"`
	Name             string             `json:"name"      bson:"name"`
	Owner            string             `json:"owner,omitempty"      bson:"owner,omitempty"`
	Protectors       []GymProtector     `json:"protectors"      bson:"protectors"`
	WasRewardClaimed bool               `json:"was_reward_claimed"      bson:"was_reward_claimed"`
}

type Item struct {
	Id                    primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Name                  string             `json:"name"      bson:"name"`
	Serial                int                `json:"serial" bson:"serial"`
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
	Serial                int                `json:"serial" bson:"serial"`
	EffectiveUntil        int                `json:"effective_until"      bson:"effective_until"`
	DecayUntil            int                `json:"decay_until"      bson:"decay_until"`
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
	LoomieTeam                      []primitive.ObjectID `json:"loomie_team"   bson:"loomie_team"`
	IsVerified                      bool                 `json:"isVerified"   bson:"isVerified"`
	CurrentLoomiesGenerationTimeout int64                `json:"currentLoomiesGenerationTimeout"   bson:"currentLoomiesGenerationTimeout"`
	LastLoomieGenerationTime        int64                `json:"lastLoomieGenerationTime"   bson:"lastLoomieGenerationTime"`
}

type PopulatedIventoryItem struct {
	Id             primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Name           string             `json:"name"      bson:"name"`
	Description    string             `json:"description"     bson:"description"`
	Target         string             `json:"target"  bson:"target"`
	Is_combat_item bool               `json:"is_combat_item"   bson:"is_combat_item"`
	Quantity       int                `json:"quantity" bson:"quantity"`
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
	Id                     primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Serial                 int                  `json:"serial"      bson:"serial"`
	Name                   string               `json:"name"      bson:"name"`
	Types                  []primitive.ObjectID `json:"types"     bson:"types"`
	Rarity                 primitive.ObjectID   `json:"rarity"     bson:"rarity"`
	HP                     int                  `json:"hp"     bson:"hp"`
	Attack                 int                  `json:"attack"     bson:"attack"`
	Defense                int                  `json:"defense"     bson:"defense"`
	ZoneId                 primitive.ObjectID   `json:"zone_id"     bson:"zone_id"`
	Latitude               float64              `json:"latitude"     bson:"latitude"`
	Longitude              float64              `json:"longitude"     bson:"longitude"`
	GeneratedAt            int64                `json:"generated_at"     bson:"generated_at"`
	Level                  int                  `json:"level"     bson:"level"`
	Experience             float64              `json:"experience"     bson:"experience"`
	UsersAlreadyCapturedIt []primitive.ObjectID `json:"users_already_captured_it"     bson:"users_already_captured_it"`
}

type AuthenticationCode struct {
	Id        primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Type      string             `json:"type"      bson:"type"`
	Email     string             `json:"email"      bson:"email"`
	Code      string             `json:"code"      bson:"code"`
	ExpiresAt int64              `json:"expires_at"      bson:"expires_at"`
}

type ValidationCode struct {
	Email             string `json:"email"`
	ValidationCode    string `json:"validationCode"`
	ValidationCodeExp int64  `json:"validationCodeExp,omitempty"`
}

type ResetPasswordCode struct {
	ResetPassCode    string `json:"resetPassCode"`
	ResetPassCodeExp int64  `json:"resetPassCodeExp,omitempty"`
	Email            string `json:"email"`
	Password         string `json:"password"  bson:"password"`
}

type CaughtLoomie struct {
	Owner      primitive.ObjectID   `json:"owner,omitempty"       bson:"owner,omitempty"`
	IsBusy     bool                 `json:"is_busy"      bson:"is_busy"`
	Id         primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Serial     int                  `json:"serial"      bson:"serial"`
	Name       string               `json:"name"      bson:"name"`
	Types      []primitive.ObjectID `json:"types"     bson:"types"`
	Rarity     primitive.ObjectID   `json:"rarity"     bson:"rarity"`
	HP         int                  `json:"hp"     bson:"hp"`
	Attack     int                  `json:"attack"     bson:"attack"`
	Defense    int                  `json:"defense"     bson:"defense"`
	Level      int                  `json:"level"     bson:"level"`
	Experience float64              `json:"experience"     bson:"experience"`
}

// ToGymProtector Converts a caught loomie to a gym protector keeping only the relevant fields
func (caughtLoomie *CaughtLoomie) ToGymProtector() *GymProtector {
	return &GymProtector{
		Serial: caughtLoomie.Serial,
		Name:   caughtLoomie.Name,
		Level:  caughtLoomie.Level,
	}
}

// ToPopulatedGym Converts the database response to a populated gym
func (aux *PopulatedGymAux) ToPopulatedGym() *PopulatedGym {
	// Remove unneded fields form the loomies
	var loomies []GymProtector = []GymProtector{}
	populatedGym := PopulatedGym{
		Id:   aux.Id,
		Name: aux.Name,
	}

	for _, loomie := range aux.Protectors {
		loomies = append(loomies, *loomie.ToGymProtector())
	}

	populatedGym.Protectors = loomies

	if len(aux.Owner) == 1 {
		populatedGym.Owner = aux.Owner[0].Username
	}

	return &populatedGym
}
