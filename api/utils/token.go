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
var JWTKeyWS = configuration.GetWsTokenSecret()

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

func CreateWsToken(userID string) (string, error) {
	// 5 minutes short lived token
	wsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userid":    userID,
		"notBefore": time.Now(),
		"expire":    time.Now().Add(time.Minute * 5),
	})

	// sign with secret and get encoded token
	var err error
	wsTokenString, err := wsToken.SignedString([]byte(JWTKeyWS))
	if err != nil {
		return "", errors.New("Could not create websocket token")
	}

	return wsTokenString, nil
}

// ValidateRefreshToken validates the refresh token is valid and not expired and returns the user id
func ValidateAccessToken(accessToken string) (string, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}
		return []byte(JWTKeyAC), nil
	})

	if err != nil {
		return "", errors.New("Invalid access token (parse)")
	}

	// validate token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, err := time.Parse(time.RFC3339, claims["expire"].(string))

		if err != nil {
			return "", errors.New("Invalid access token (expire format)")
		}

		if exp.Before(time.Now()) {
			return "", errors.New("Access token expired")
		}

		return claims["userid"].(string), nil
	} else {
		return "", errors.New("Invalid access token (claims)")
	}
}

// ValidateRefreshToken validates the refresh token is valid and not expired and returns the user id
func ValidateRefreshToken(refreshToken string) (string, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}
		return []byte(JWTKeyRF), nil
	})

	if err != nil {
		return "", errors.New("Invalid refresh token (parse)")
	}

	// validate token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, err := time.Parse(time.RFC3339, claims["expire"].(string))

		if err != nil {
			return "", errors.New("Invalid refresh token (expire format)")
		}

		if exp.Before(time.Now()) {
			return "", errors.New("Refresh token expired")
		}

		return claims["userid"].(string), nil
	} else {
		return "", errors.New("Invalid refresh token (claims)")
	}
}
