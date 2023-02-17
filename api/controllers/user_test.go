package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/tests"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// ### Global variables ###
var usersCollection = configuration.ConnectToMongoCollection("users")

// ### Tests ###
// TestSignupSuccessAndConflict tests the signup endpoint with a valid payload and a payload with an already existing email / username
func TestSignupSuccessAndConflict(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create random payload
	payload := tests.GenerateRandomUser()

	// Setup router to create the requests
	router := tests.SetupGinRouter()
	router.POST("/signup", HandleSignUp)

	// Make the request and get the JSON response
	w, req := tests.SetupPostRequest("/signup", payload)
	_, err := ioutil.ReadAll(w.Body)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	// 1. Success request
	c.NoError(err)
	c.Equal(http.StatusOK, w.Code)
	c.Equal("User created successfully", response["message"])

	// Check if user was created in the database
	var user interfaces.User
	err = usersCollection.FindOne(ctx, bson.D{{Key: "email", Value: payload.Email}}).Decode(&user)
	c.NoError(err)
	c.Equal(payload.Email, user.Email)
	c.Equal(payload.Username, user.Username)

	// 2. Conflict request with the same email
	w, req = tests.SetupPostRequest("/signup", payload)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.NoError(err)
	c.Equal(http.StatusConflict, w.Code)
	c.Equal("Email already exists", response["message"])

	// 3. Conflict request with the same username
	oldEmail := payload.Email
	payload.Email = tests.FakerInstance.Internet().Email()
	w, req = tests.SetupPostRequest("/signup", payload)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.NoError(err)
	c.Equal(http.StatusConflict, w.Code)
	c.Equal("Username already exists", response["message"])

	// Delete users
	_, err = usersCollection.DeleteOne(ctx, bson.D{{Key: "email", Value: oldEmail}})
	c.NoError(err)
	_, err = usersCollection.DeleteOne(ctx, bson.D{{Key: "email", Value: payload.Email}})
	c.NoError(err)
}

// TestSignupSuccess tests the signup endpoint with a non verified user
func TestLoginForbidden(t *testing.T) {
	c := require.New(t)
	var response map[string]string
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	router.POST("/signup", HandleSignUp)
	w, req := tests.SetupPostRequest("/signup", randomUser)
	router.ServeHTTP(w, req)

	// Try to login with the random user
	loginForm := map[string]string{
		"email":    randomUser.Email,
		"password": randomUser.Password,
	}
	router.POST("/login", HandleLogIn)
	w, req = tests.SetupPostRequest("/login", loginForm)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check if the response is correct
	c.Equal(http.StatusForbidden, w.Code)
	c.Equal("User has not been verified", response["message"])

	// Delete the user from the database
	usersCollection.DeleteOne(ctx, map[string]string{"email": randomUser.Email})
}
