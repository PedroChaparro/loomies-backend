package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/jaswdr/faker"
)

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

// SetupPostRequest creates a new POST request with the given payload and headers (if any)
func SetupPostRequest(endpoint string, payload interface{}, headers ...CustomHeader) (*httptest.ResponseRecorder, *http.Request) {
	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewReader(payloadBytes))
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
	router.POST("/signup", handler)
	w, req := SetupPostRequest("/signup", user)
	router.ServeHTTP(w, req)
}
