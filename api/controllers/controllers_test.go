package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/PedroChaparro/loomies-backend/tests"
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

	// Verify the user
	usersCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "isVerified", Value: true}}}})

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
	c.Equal(http.StatusOK, w.Code)
	c.Equal("Successfully logged in", response["message"])
	c.NotEmpty(response["accessToken"])
	c.NotEmpty(response["refreshToken"])

	// Delete the user from the database
	usersCollection.DeleteOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}})
}
