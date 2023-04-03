package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/tests"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// ### Global variables ###
var usersCollection = configuration.ConnectToMongoCollection("users")
var authenticationCodesCollection = configuration.ConnectToMongoCollection("authentication_codes")

// ### Tests ###
// TestSignupSuccessAndConflict Tests the signup endpoint with a valid payload and a payload with an already existing email / username
func TestSignupSuccessAndConflict(t *testing.T) {
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create random payload
	payload := tests.GenerateRandomUser()

	// Setup router to create the requests
	router := tests.SetupGinRouter()
	router.POST("/user/signup", HandleSignUp)

	// Make the request and get the JSON response
	w, req := tests.SetupPostRequest("/user/signup", payload)
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
	w, req = tests.SetupPostRequest("/user/signup", payload)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.NoError(err)
	c.Equal(http.StatusConflict, w.Code)
	c.Equal("Email already exists", response["message"])

	// 3. Conflict request with the same username
	oldEmail := payload.Email
	payload.Email = tests.FakerInstance.Internet().Email()
	w, req = tests.SetupPostRequest("/user/signup", payload)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.NoError(err)
	c.Equal(http.StatusConflict, w.Code)
	c.Equal("Username already exists", response["message"])

	// Delete users
	err = tests.DeleteUser(oldEmail)
	c.NoError(err)
	err = tests.DeleteUser(payload.Email)
	c.NoError(err)
}

// TestAccountValidationCodeSuccess Test the success case for /user/validate/code endpoint
func TestAccountValidationCodeSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Make the request and get the JSON response
	router.POST("/user/validate/code", HandleAccountValidationCodeRequest)
	w, req := tests.SetupPostRequest("/user/validate/code", map[string]string{"email": randomUser.Email})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the response
	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal("New Code created and sended", response["message"])

	// Delete user
	err := tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}

// TestAccountValidationCodeBadRequest Test the Bad Request cases for /user/validate/code endpoint
func TestAccountValidationCodeBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// 1. Test without JSON payload
	router.POST("/user/validate/code", HandleAccountValidationCodeRequest)
	w, req := tests.SetupPostRequest("/user/validate/code", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Bad request", response["message"])

	// 2. Test with empty email
	w, req = tests.SetupPostRequest("/user/validate/code", map[string]string{"email": ""})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Email cannot be empty", response["message"])

	// 3. Test with unregistred email
	w, req = tests.SetupPostRequest("/user/validate/code", map[string]string{"email": "unexisting@gmail.com"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusNotFound, w.Code)
	c.Equal(true, response["error"])
	c.Equal("This Email has not been registered", response["message"])

	// 4. Test with verified email
	usersCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "isVerified", Value: true}}}})
	w, req = tests.SetupPostRequest("/user/validate/code", map[string]string{"email": randomUser.Email})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusConflict, w.Code)
	c.Equal(true, response["error"])
	c.Equal("This Email has been already verified", response["message"])

	// Delete user
	err := tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}

// TestAccountValidationSuccess Test the success case for /user/validate endpoint
func TestAccountValidationSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Get the code from the database
	var code interfaces.AuthenticationCode
	err := authenticationCodesCollection.FindOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}, {Key: "type", Value: "ACCOUNT_VERIFICATION"}}).Decode(&code)
	c.NoError(err)

	// Make the request and get the JSON response
	router.POST("/user/validate", HandleAccountValidation)
	w, req := tests.SetupPostRequest("/user/validate", map[string]string{"email": randomUser.Email, "validationCode": code.Code})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the response
	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Email has been verified", response["message"])
}

// TestAccountValidationBadRequest Test the Bad Request cases for /user/validate endpoint
func TestAccountValidationBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Get the code from the database
	var code interfaces.AuthenticationCode
	err := authenticationCodesCollection.FindOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}, {Key: "type", Value: "ACCOUNT_VERIFICATION"}}).Decode(&code)
	c.NoError(err)
	codeNumber, _ := strconv.Atoi(code.Code)

	// 1. Test without JSON payload
	router.POST("/user/validate", HandleAccountValidation)
	w, req := tests.SetupPostRequest("/user/validate", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Bad request", response["message"])

	// 2. Test with empty email
	w, req = tests.SetupPostRequest("/user/validate", map[string]string{"email": "", "validationCode": code.Code})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Email cannot be empty", response["message"])

	// 3. Test with empty validationCode
	w, req = tests.SetupPostRequest("/user/validate", map[string]string{"email": randomUser.Email, "validationCode": ""})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Verification code cannot be empty", response["message"])

	// 4. Test with incorrect validationCode
	w, req = tests.SetupPostRequest("/user/validate", map[string]string{"email": randomUser.Email, "validationCode": strconv.Itoa(codeNumber + 1)})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusUnauthorized, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Code was incorrect or time has expired", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}
