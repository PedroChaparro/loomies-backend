package interfaces

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Globals struct {
	Loaded             bool
	MongoClient        *mongo.Client
	AccessTokenSecret  string
	RefreshTokenSecret string
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
