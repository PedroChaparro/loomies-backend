package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// ### Global variables ###
var usersCollection = configuration.ConnectToMongoCollection("users")
var fake = faker.New()

// ### Helper functions ###
// SetupGinRouter creates a new gin router to be used in the tests
func setupGinRouter() *gin.Engine {
	router := gin.Default()
	return router
}

// SetupPostRequest creates a new POST request with the given payload
func setupPostRequest(endpoint string, payload interface{}) (*httptest.ResponseRecorder, *http.Request) {
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	return w, req
}

// GenerateRandomUser generates a random and valid user
func generateRandomUser() interfaces.User {
	var user interfaces.User
	user.Email = fake.Internet().Email()
	user.Password = fake.Internet().Password() + "A1!" // Adding uppercase, number and special character
	user.Username = fake.Internet().User()
	return user
}

// ### Tests ###
// TestSignupSuccessAndConflict tests the signup endpoint with a valid payload and a payload with an already existing email / username
func TestSignupSuccessAndConflict(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create random payload
	payload := generateRandomUser()

	// Setup router to create the requests
	router := setupGinRouter()
	router.POST("/signup", HandleSignUp)

	// Make the request and get the JSON response
	w, req := setupPostRequest("/signup", payload)
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
	w, req = setupPostRequest("/signup", payload)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.NoError(err)
	c.Equal(http.StatusConflict, w.Code)
	c.Equal("Email already exists", response["message"])

	// 3. Conflict request with the same username
	oldEmail := payload.Email
	payload.Email = fake.Internet().Email()
	w, req = setupPostRequest("/signup", payload)
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
