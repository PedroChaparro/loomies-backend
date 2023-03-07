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

// todo req validation code, interface
type ValidationCode struct {
	Email          string `json:"email"`
	ValidationCode string `json:"validationCode"`
}
