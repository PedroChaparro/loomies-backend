package interfaces

import "go.mongodb.org/mongo-driver/bson/primitive"

type SignUpForm struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogInForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmailForm struct {
	Email string `json:"email"`
}

type ClaimGymRewardReq struct {
	GymID     string  `json:"gym_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RegisterCombatReq struct {
	GymID     string  `json:"gym_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CatchLoomieForm struct {
	LoomieId   string  `json:"loomie_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	LoomballId string  `json:"loomball_id"`
}

type FuseLoomiesReq struct {
	LoomieId1 string `json:"loomie_id_1"`
	LoomieId2 string `json:"loomie_id_2"`
}

type UpdateLoomieTeamReq struct {
	LoomieTeam []string `json:"loomie_team"`
}

type UseNotCombatItemReq struct {
	ItemId   string `json:"item_id"`
	LoomieId string `json:"loomie_id"`
}

type PopulatedInventoryItem struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id"`
	Name     string             `json:"name" bson:"name"`
	Serial   int                `json:"serial" bson:"serial"`
	Quantity int                `json:"quantity" bson:"quantity"`
}
