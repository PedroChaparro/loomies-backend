package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// secret keys
var JWTKeyAC = []byte(os.Getenv("SECRET_AC"))
var JWTKeyRF = []byte(os.Getenv("SECRET_RF"))

func CreateAccessToken(userID string) (string, error) {
	// 30 minutes short lived token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid":    userID,
		"notBefore": time.Now(),
		"expire":    time.Now().Add(time.Minute * 30),
	})

	// sign with secret and get encoded token
	var err error
	accessTokenString, err := accessToken.SignedString(JWTKeyAC)
	if err != nil {
		return "", errors.New("Could not create access token")
	}

	return accessTokenString, nil
}

func CreateRefreshToken(userID string) (string, error) {
	// 5 months long lived token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid":    userID,
		"notBefore": time.Now(),
		"expire":    time.Now().AddDate(0, 5, 0),
	})

	// sign with secret and get encoded token
	var err error
	refreshTokenString, err := refreshToken.SignedString(JWTKeyRF)
	if err != nil {
		return "", errors.New("Could not create refresh token")
	}

	return refreshTokenString, nil
}
