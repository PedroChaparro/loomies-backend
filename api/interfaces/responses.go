package interfaces

import "go.mongodb.org/mongo-driver/bson/primitive"

// NearGymsRes is the response for the near gyms endpoint
// It's a subset of the Gym struct to remove unnecessary fields
type NearGymsRes struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	Latitude  float64            `json:"latitude"      bson:"latitude"`
	Longitude float64            `json:"longitude"      bson:"longitude"`
	Name      string             `json:"name"      bson:"name"`
	Owner     primitive.ObjectID `json:"owner,omitempty"      bson:"owner,omitempty"`
}

// ToNearGymsRes converts a Gym struct to a NearGymsRes struct
func (res *Gym) ToNearGymsRes() *NearGymsRes {
	return &NearGymsRes{
		Id:        res.Id,
		Latitude:  res.Latitude,
		Longitude: res.Longitude,
		Name:      res.Name,
		Owner:     res.Owner,
	}
}
