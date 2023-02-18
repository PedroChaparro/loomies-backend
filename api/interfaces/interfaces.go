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

type Zone struct {
	Id             primitive.ObjectID `json:"_id" bson:"_id"`
	LeftFrontier   float64            `json:"leftFrontier" bson:"leftFrontier"`
	RightFrontier  float64            `json:"rightFrontier" bson:"rightFrontier"`
	TopFrontier    float64            `json:"topFrontier" bson:"topFrontier"`
	BottomFrontier float64            `json:"bottomFrontier" bson:"bottomFrontier"`
	Number         int                `json:"number" bson:"number"`
	Gym            primitive.ObjectID `json:"gym" bson:"gym"`
}

type User struct {
	Id         primitive.ObjectID   `json:"_id,omitempty"       bson:"_id,omitempty"`
	Username   string               `json:"username"      bson:"username"`
	Email      string               `json:"email"     bson:"email"`
	Password   string               `json:"password"  bson:"password"`
	Items      []interface{}        `json:"items"     bson:"items"`
	Loomies    []primitive.ObjectID `json:"loomies"   bson:"loomies"`
	IsVerified bool                 `json:"isVerified"   bson:"isVerified"`
}
