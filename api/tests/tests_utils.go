package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/jaswdr/faker"
	"go.mongodb.org/mongo-driver/bson"
)

var usersCollection = configuration.ConnectToMongoCollection("users")
var authenticationCodesCollection = configuration.ConnectToMongoCollection("authentication_codes")

// ### Types / Structs
type CustomHeader struct {
	Name  string
	Value string
}

// Fake: faker instance to generate random data
var FakerInstance = faker.New()

// SetupGinRouter creates a new gin router to be used in the tests
func SetupGinRouter() *gin.Engine {
	router := gin.Default()
	return router
}

// SetupPayloadedRequest creates a new POST or PUT request with the given payload and headers (if any)
func SetupPayloadedRequest(endpoint string, method string, payload interface{}, headers ...CustomHeader) (*httptest.ResponseRecorder, *http.Request) {
	var req *http.Request

	if payload != nil {
		payloadBytes, _ := json.Marshal(payload)
		r, _ := http.NewRequest(method, endpoint, bytes.NewReader(payloadBytes))
		req = r
		req.Header.Set("Content-Type", "application/json")
	} else {
		r, _ := http.NewRequest(method, endpoint, nil)
		req = r
	}

	// Set the custom headers
	if len(headers) > 0 {
		for _, header := range headers {
			req.Header.Set(header.Name, header.Value)
		}
	}

	w := httptest.NewRecorder()
	return w, req
}

// SetupGetRequest creates a new GET request with the given headers (if any)
func SetupGetRequest(endpoint string, headers ...CustomHeader) (*httptest.ResponseRecorder, *http.Request) {
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Content-Type", "application/json")

	// Set the custom headers
	if len(headers) > 0 {
		for _, header := range headers {
			req.Header.Set(header.Name, header.Value)
		}
	}

	w := httptest.NewRecorder()
	return w, req
}

// GenerateRandomUser generates a random and valid user
func GenerateRandomUser() interfaces.User {
	var user interfaces.User
	user.Email = FakerInstance.Internet().Email()
	user.Password = FakerInstance.Internet().Password() + "A1!" // Adding uppercase, number and special character
	user.Username = FakerInstance.Internet().User()
	return user
}

// InsertUser inserts a random user in the database. Use this function when you need to test
// endpoints that require a user to be logged in whintout having to test the signup endpoint
func InsertUser(user interfaces.User, router *gin.Engine, handler gin.HandlerFunc) {
	router.POST("/user/signup", handler)
	w, req := SetupPayloadedRequest("/user/signup", "POST", user)
	router.ServeHTTP(w, req)
}

// DeleteUser Deletes a user from the database
func DeleteUser(email string) error {
	_, err := usersCollection.DeleteOne(context.Background(), bson.D{{Key: "email", Value: email}})

	if err != nil {
		return err
	}

	_, err = authenticationCodesCollection.DeleteMany(context.Background(), bson.D{{Key: "email", Value: email}})
	return err
}
