package interfaces

import "time"

type SignUpForm struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogInForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ValidationCode struct {
	Email          string    `json:"email"`
	ValidationCode string    `json:"validationCode"`
	TimeExpiration time.Time `json:"timeExpiration"`
}

type Email struct {
	Email string `json:"email"`
}

type ClaimGymRewardReq struct {
	GymID     string  `json:"gym_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
