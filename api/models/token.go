package models

import (
	"errors"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/golang-jwt/jwt/v4"
)

// secret key

var JWTKey = []byte("my_secret_key")

// create jwt, accessToken and refreshToken

func CreateToken(info interfaces.TokenInfo) (string, string, *jwt.Token, *jwt.Token, error) {

	// 1 month
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid":    info.UserID,
		"notBefore": time.Now(),
		"expire":    time.Now().Add(time.Hour * 720),
	})

	// 31 day
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid":    info.UserID,
		"notBefore": time.Now(),
		"expire":    time.Now().AddDate(0, 0, 31),
	})

	// sign with secret
	// and get encoded token

	var err error
	accessTokenString, err := accessToken.SignedString(JWTKey)
	if err != nil {
		return "", "", nil, nil, errors.New("Could not create access token")
	}

	refreshTokenString, err := refreshToken.SignedString(JWTKey)
	if err != nil {
		return "", "", nil, nil, errors.New("Could not create refresh token")
	}

	return accessTokenString, refreshTokenString, accessToken, refreshToken, nil
}
