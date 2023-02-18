package utils

import (
	"errors"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/golang-jwt/jwt/v4"
)

// secret keys
var JWTKeyAC = configuration.GetAccessTokenSecret()
var JWTKeyRF = configuration.GetRefreshTokenSecret()

func CreateAccessToken(userID string) (string, error) {
	// 30 minutes short lived token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid":    userID,
		"notBefore": time.Now(),
		"expire":    time.Now().Add(time.Minute * 30),
	})

	// sign with secret and get encoded token
	var err error
	accessTokenString, err := accessToken.SignedString([]byte(JWTKeyAC))
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
	refreshTokenString, err := refreshToken.SignedString([]byte(JWTKeyRF))
	if err != nil {
		return "", errors.New("Could not create refresh token")
	}

	return refreshTokenString, nil
}
