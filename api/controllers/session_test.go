package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/PedroChaparro/loomies-backend/tests"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// ## Helper functions

// loginWithRandomUser creates a random user, inserts it into the database, verifies it and tries to login with it
func loginWithRandomUser() (interfaces.User, map[string]string) {
	ctx := context.Background()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Verify the user and save the database document
	var databaseUser interfaces.User
	usersCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "isVerified", Value: true}}}})
	usersCollection.FindOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}).Decode(&databaseUser)

	// Try to login with the random user
	var response map[string]string
	loginForm := map[string]string{
		"email":    randomUser.Email,
		"password": randomUser.Password,
	}

	router.POST("/session/login", HandleLogIn)
	w, req := tests.SetupPostRequest("/session/login", loginForm)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	return databaseUser, response
}

// ## Tests

// TestSignupSuccess tests the signup endpoint with a non verified user
func TestLoginForbidden(t *testing.T) {
	c := require.New(t)
	var response map[string]string
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Try to login with the random user
	loginForm := map[string]string{
		"email":    randomUser.Email,
		"password": randomUser.Password,
	}

	router.POST("/session/login", HandleLogIn)
	w, req := tests.SetupPostRequest("/session/login", loginForm)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check if the response is correct
	c.Equal(http.StatusForbidden, w.Code)
	c.Equal("User has not been verified", response["message"])

	// Delete the user from the database
	usersCollection.DeleteOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}})
}

// TestSignupSuccess tests the signup endpoint with a verified user
func TestLoginSuccess(t *testing.T) {
	c := require.New(t)
	var response map[string]string
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Verify the user and save the database document
	usersCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "isVerified", Value: true}}}})
	var databaseUser interfaces.User
	usersCollection.FindOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}).Decode(&databaseUser)

	// Try to login with the random user
	loginForm := map[string]string{
		"email":    randomUser.Email,
		"password": randomUser.Password,
	}

	router.POST("/session/login", HandleLogIn)
	w, req := tests.SetupPostRequest("/session/login", loginForm)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// 1. Check if the response fields are correct
	c.Equal(http.StatusOK, w.Code)
	c.Equal("Successfully logged in", response["message"])
	c.NotEmpty(response["accessToken"])
	c.NotEmpty(response["refreshToken"])

	// 2. Check tokens claims
	accessTokenClaims := jwt.MapClaims{}
	refreshTokenClaims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(response["accessToken"], accessTokenClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(configuration.GetAccessTokenSecret()), nil
	})
	c.NoError(err)

	_, err = jwt.ParseWithClaims(response["refreshToken"], refreshTokenClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(configuration.GetRefreshTokenSecret()), nil
	})
	c.NoError(err)

	c.Equal(databaseUser.Id.Hex(), accessTokenClaims["userid"])
	c.Equal(databaseUser.Id.Hex(), refreshTokenClaims["userid"])

	// Delete the user from the database
	usersCollection.DeleteOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}})
}

// TestRefreshUnauthorized tests the refresh endpoint without a refresh token
func TestRefreshUnauthorized(t *testing.T) {
	c := require.New(t)
	router := tests.SetupGinRouter()

	// Try to refresh without a refresh token
	var refreshResponse map[string]string
	router.POST("/session/refresh", middlewares.MustProvideRefreshToken(), HandleRefresh)
	w, req := tests.SetupPostRequest("/session/refresh", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &refreshResponse)

	// Check if the response is correct
	c.Equal(http.StatusUnauthorized, w.Code)
	c.Equal("Refresh token is required", refreshResponse["message"])
}

// TestRefreshSuccess tests the refresh endpoint with a valid refresh token
func TestRefreshSuccess(t *testing.T) {
	c := require.New(t)
	router := tests.SetupGinRouter()
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	databaseUser, loginResponse := loginWithRandomUser()

	// Get a new access token from the refresh token
	var refreshResponse map[string]string
	router.POST("/session/refresh", middlewares.MustProvideRefreshToken(), HandleRefresh)
	w, req := tests.SetupPostRequest("/session/refresh", nil, tests.CustomHeader{Name: "Refresh-Token", Value: loginResponse["refreshToken"]})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &refreshResponse)

	// 1. Check if the response fields are correct
	c.Equal(http.StatusOK, w.Code)
	c.NotEmpty(refreshResponse["accessToken"])
	c.Equal("Successfully refreshed access token", refreshResponse["message"])

	// 2. Check tokens claims
	accessTokenClaims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(refreshResponse["accessToken"], accessTokenClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(configuration.GetAccessTokenSecret()), nil
	})

	c.NoError(err)
	c.Equal(databaseUser.Id.Hex(), accessTokenClaims["userid"])

	// Remove the user from the database
	usersCollection.DeleteOne(ctx, bson.D{{Key: "email", Value: databaseUser.Email}})
}

// TestWhoamiUnauthorized tests the whoami endpoint without a valid access token
func TestWhoamiUnauthorized(t *testing.T) {
	c := require.New(t)
	router := tests.SetupGinRouter()

	// Try to get the user without a valid access token
	var whoamiResponse map[string]string
	router.GET("/session/whoami", middlewares.MustProvideAccessToken(), HandleWhoami)
	w, req := tests.SetupGetRequest("/session/whoami")
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &whoamiResponse)

	// Check if the response is correct
	c.Equal(http.StatusUnauthorized, w.Code)
	c.Equal("Access token is required", whoamiResponse["message"])
}

// TestWhoamiSuccess tests the whoami endpoint with a valid access token
func TestWhoamiSuccess(t *testing.T) {
	c := require.New(t)
	router := tests.SetupGinRouter()
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	databaseUser, loginResponse := loginWithRandomUser()

	// Make a request to the whoami endpoint
	var whoamiResponse map[string]interface{}
	router.GET("/session/whoami", middlewares.MustProvideAccessToken(), HandleWhoami)
	w, req := tests.SetupGetRequest("/session/whoami", tests.CustomHeader{Name: "Access-Token", Value: loginResponse["accessToken"]})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &whoamiResponse)

	// 1. Check if the response fields are correct
	c.Equal(http.StatusOK, w.Code)
	c.Equal("Successfully retrieved user", whoamiResponse["message"])

	whoamiResponseUser := whoamiResponse["user"].(map[string]interface{})
	c.Equal(databaseUser.Email, whoamiResponseUser["email"])
	c.Equal(databaseUser.Username, whoamiResponseUser["username"])

	// Remove the user from the database
	usersCollection.DeleteOne(ctx, bson.D{{Key: "email", Value: databaseUser.Email}})
}
