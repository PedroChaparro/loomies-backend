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
func SetupGinRouter() *gin.Engine {
	router := gin.Default()
	return router
}

// SetupPostRequest creates a new POST request with the given payload
func SetupPostRequest(endpoint string, payload interface{}) (*httptest.ResponseRecorder, *http.Request) {
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	return w, req
}

// ### Tests ###
// TestSignupSuccess tests the signup endpoint with a valid payload
func TestSignupSuccess(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create random payload
	var payload interfaces.SignUpForm
	payload.Email = fake.Internet().Email()
	payload.Password = fake.Internet().Password() + "A1!" // Adding uppercase, number and special character
	payload.Username = fake.Internet().User()

	// Setup router to create the requests
	router := SetupGinRouter()
	router.POST("/signup", HandleSignUp)

	// Make the request and get the JSON response
	w, req := SetupPostRequest("/signup", payload)
	_, err := ioutil.ReadAll(w.Body)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the response
	c.NoError(err)
	c.Equal(http.StatusOK, w.Code)
	c.Equal("User created successfully", response["message"])

	// Check if user was created in the database
	var user interfaces.User
	err = usersCollection.FindOne(ctx, bson.D{{Key: "email", Value: payload.Email}}).Decode(&user)
	c.NoError(err)
	c.Equal(payload.Email, user.Email)
	c.Equal(payload.Username, user.Username)

	// Delete user
	_, err = usersCollection.DeleteOne(ctx, bson.D{{Key: "email", Value: payload.Email}})
	c.NoError(err)
}
