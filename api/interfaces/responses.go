package interfaces

import "go.mongodb.org/mongo-driver/bson/primitive"

// NearGymsRes is the response for the near gyms endpoint
// It's a subset of the Gym struct to remove unnecessary fields
type NearGymsRes struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	Latitude  float64            `json:"latitude"      bson:"latitude"`
	Longitude float64            `json:"longitude"      bson:"longitude"`
	Name      string             `json:"name"      bson:"name"`
}

// LoomieRarityRes just bring the name field of lommie rarity
type LoomieRarityRes struct {
	Name string `json:"name"      bson:"name"`
}

// UserItemsRes is the response for the /items endpoint to avoid sending unnecessary
// data to the client
type UserItemsRes struct {
	Id             primitive.ObjectID `json:"_id" bson:"_id"`
	Name           string             `json:"name"      bson:"name"`
	Description    string             `json:"description"      bson:"description"`
	Target         string             `json:"target"      bson:"target"`
	Is_combat_item bool               `json:"is_combat_item"      bson:"is_combat_item"`
	Quantity       int                `json:"quantity"      bson:"quantity"`
}

// UserLoomballsRes is the response for the /loomballs endpoint to avoid sending unnecessary
// data to the client
type UserLoomballsRes struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id"`
	Name     string             `json:"name"      bson:"name"`
	Quantity int                `json:"quantity"      bson:"quantity"`
}

// Show info of user loomies
type UserLoomiesRes struct {
	Id      primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Serial  int                `json:"serial"      bson:"serial"`
	Name    string             `json:"name"      bson:"name"`
	Types   []string           `json:"types"     bson:"types"`
	Rarity  string             `json:"rarity"     bson:"rarity"`
	Hp      int                `json:"hp"     bson:"hp"`
	Attack  int                `json:"attack"     bson:"attack"`
	IsBusy  bool               `json:"is_busy"     bson:"is_busy"`
	Defense int                `json:"defense"     bson:"defense"`
}

// This is an aux for show info user loomies
type UserLoomiesResAux struct {
	Id      primitive.ObjectID `json:"_id,omitempty"       bson:"_id,omitempty"`
	Serial  int                `json:"serial"      bson:"serial"`
	Name    string             `json:"name"      bson:"name"`
	Types   []LoomieType       `json:"types"     bson:"types"`
	Rarity  []LoomieRarityRes  `json:"rarity"     bson:"rarity"`
	Hp      int                `json:"hp"     bson:"hp"`
	Attack  int                `json:"attack"     bson:"attack"`
	IsBusy  bool               `json:"is_busy"     bson:"is_busy"`
	Defense int                `json:"defense"     bson:"defense"`
}

// ToNearGymsRes converts a Gym struct to a NearGymsRes struct
func (res *Gym) ToNearGymsRes() *NearGymsRes {
	return &NearGymsRes{
		Id:        res.Id,
		Latitude:  res.Latitude,
		Longitude: res.Longitude,
		Name:      res.Name,
	}
}
