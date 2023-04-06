package interfaces

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
