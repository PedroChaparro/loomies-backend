package utils

import (
	"errors"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
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

func CreateWsToken(userID string, gymId string, latitude float64, longitude float64) (string, error) {
	wsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"gym_id":    gymId,
		"latitude":  latitude,
		"longitude": longitude,
		"notBefore": time.Now(),
		"expire":    time.Now().Add(time.Minute * 1),
	})

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

// ValidateWsToken validates the websocket token is valid and not expired and returns the token claims
func ValidateWsToken(wsToken string) (interfaces.WsTokenClaims, error) {
	token, err := jwt.Parse(wsToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method")
		}
		return []byte(JWTKeyWS), nil
	})

	if err != nil {
		return interfaces.WsTokenClaims{}, errors.New("Invalid websocket token (parse)")
	}

	// validate token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		exp, err := time.Parse(time.RFC3339, claims["expire"].(string))

		if err != nil {
			return interfaces.WsTokenClaims{}, errors.New("Invalid websocket token (expire format)")
		}

		if exp.Before(time.Now()) {
			return interfaces.WsTokenClaims{}, errors.New("Websocket token expired")
		}

		return interfaces.WsTokenClaims{
			UserID:    claims["user_id"].(string),
			GymID:     claims["gym_id"].(string),
			Latitude:  claims["latitude"].(float64),
			Longitude: claims["longitude"].(float64),
		}, nil
	} else {
		return interfaces.WsTokenClaims{}, errors.New("Invalid websocket token (claims)")
	}
}
