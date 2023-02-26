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
}

type Coordinates struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

type Zone struct {
	Id             primitive.ObjectID `json:"_id" bson:"_id"`
	LeftFrontier   float64            `json:"leftFrontier" bson:"leftFrontier"`
	RightFrontier  float64            `json:"rightFrontier" bson:"rightFrontier"`
	TopFrontier    float64            `json:"topFrontier" bson:"topFrontier"`
	BottomFrontier float64            `json:"bottomFrontier" bson:"bottomFrontier"`
	Number         int                `json:"number" bson:"number"`
	Coordinates    string             `json:"coordinates" bson:"coordinates"`
	Gym            primitive.ObjectID `json:"gym" bson:"gym"`
}

type ZoneWithGyms struct {
	Id             primitive.ObjectID `json:"_id" bson:"_id"`
	LeftFrontier   float64            `json:"leftFrontier" bson:"leftFrontier"`
	RightFrontier  float64            `json:"rightFrontier" bson:"rightFrontier"`
	TopFrontier    float64            `json:"topFrontier" bson:"topFrontier"`
	BottomFrontier float64            `json:"bottomFrontier" bson:"bottomFrontier"`
	Number         int                `json:"number" bson:"number"`
	Coordinates    string             `json:"coordinates" bson:"coordinates"`
	Gym            primitive.ObjectID `json:"gym" bson:"gym"`
	Gyms           []Gym              `json:"gyms" bson:"gyms"`
}

type Gym struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	Latitude  float64            `json:"latitude"      bson:"latitude"`
	Longitude float64            `json:"longitude"      bson:"longitude"`
	Name      string             `json:"name"      bson:"name"`
}

type User struct {
	Id       primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Username string             `json:"username"      bson:"username"`
	Email    string             `json:"email"     bson:"email"`
	Password string             `json:"password"  bson:"password"`

	// tODO: Change this to a struct
	Items []interface{} `json:"items"     bson:"items"`

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
	Id        primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Name      string               `json:"name"      bson:"name"`
	Types     []primitive.ObjectID `json:"types"     bson:"types"`
	Rarity    primitive.ObjectID   `json:"rarity"     bson:"rarity"`
	HP        int                  `json:"hp"     bson:"hp"`
	Attack    int                  `json:"attack"     bson:"attack"`
	Defense   int                  `json:"defense"     bson:"defense"`
	Latitude  float64              `json:"latitude"     bson:"latitude"`
	Longitude float64              `json:"longitude"     bson:"longitude"`
}
