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

// Fake: faker instance to generate random data
var FakerInstance = faker.New()

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

// GenerateRandomUser generates a random and valid user
func GenerateRandomUser() interfaces.User {
	var user interfaces.User
	user.Email = FakerInstance.Internet().Email()
	user.Password = FakerInstance.Internet().Password() + "A1!" // Adding uppercase, number and special character
	user.Username = FakerInstance.Internet().User()
	return user
}
