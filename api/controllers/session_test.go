package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/tests"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

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

	router.POST("/login", HandleLogIn)
	w, req := tests.SetupPostRequest("/login", loginForm)
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

	router.POST("/login", HandleLogIn)
	w, req := tests.SetupPostRequest("/login", loginForm)
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
