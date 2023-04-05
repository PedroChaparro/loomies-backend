package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/tests"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
	w, req := tests.SetupPayloadedRequest("/user/signup", "POST", payload)
	_, err := ioutil.ReadAll(w.Body)
	router.ServeHTTP(w, req)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	// -------------------------
	// 1. Success request
	// -------------------------
	c.NoError(err)
	c.Equal(http.StatusOK, w.Code)
	c.Equal("User created successfully", response["message"])

	// Check if user was created in the database
	var user interfaces.User
	err = models.UserCollection.FindOne(ctx, bson.D{{Key: "email", Value: payload.Email}}).Decode(&user)
	c.NoError(err)
	c.Equal(payload.Email, user.Email)
	c.Equal(payload.Username, user.Username)

	// -------------------------
	// 2. Conflict request with the same email
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/user/signup", "POST", payload)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.NoError(err)
	c.Equal(http.StatusConflict, w.Code)
	c.Equal("Email already exists", response["message"])

	// -------------------------
	// 3. Conflict request with the same username
	// -------------------------
	oldEmail := payload.Email
	payload.Email = tests.FakerInstance.Internet().Email()
	w, req = tests.SetupPayloadedRequest("/user/signup", "POST", payload)
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
	w, req := tests.SetupPayloadedRequest("/user/validate/code", "POST", map[string]string{"email": randomUser.Email})
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

	// -------------------------
	// 1. Test without JSON payload
	// -------------------------
	router.POST("/user/validate/code", HandleAccountValidationCodeRequest)
	w, req := tests.SetupPayloadedRequest("/user/validate/code", "POST", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Bad request", response["message"])

	// -------------------------
	// 2. Test with empty email
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/user/validate/code", "POST", map[string]string{"email": ""})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Email cannot be empty", response["message"])

	// -------------------------
	// 3. Test with unregistred email
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/user/validate/code", "POST", map[string]string{"email": "unexisting@gmail.com"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusNotFound, w.Code)
	c.Equal(true, response["error"])
	c.Equal("This Email has not been registered", response["message"])

	// -------------------------
	// 4. Test with verified email
	// -------------------------
	models.UserCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "isVerified", Value: true}}}})
	w, req = tests.SetupPayloadedRequest("/user/validate/code", "POST", map[string]string{"email": randomUser.Email})
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
	err := models.AuthenticationCodesCollection.FindOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}, {Key: "type", Value: "ACCOUNT_VERIFICATION"}}).Decode(&code)
	c.NoError(err)

	// Make the request and get the JSON response
	router.POST("/user/validate", HandleAccountValidation)
	w, req := tests.SetupPayloadedRequest("/user/validate", "POST", map[string]string{"email": randomUser.Email, "validationCode": code.Code})
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
	err := models.AuthenticationCodesCollection.FindOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}, {Key: "type", Value: "ACCOUNT_VERIFICATION"}}).Decode(&code)
	c.NoError(err)
	codeNumber, _ := strconv.Atoi(code.Code)

	// -------------------------
	// 1. Test without JSON payload
	// -------------------------
	router.POST("/user/validate", HandleAccountValidation)
	w, req := tests.SetupPayloadedRequest("/user/validate", "POST", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Bad request", response["message"])

	// -------------------------
	// 2. Test with empty email
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/user/validate", "POST", map[string]string{"email": "", "validationCode": code.Code})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Email cannot be empty", response["message"])

	// -------------------------
	// 3. Test with empty validationCode
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/user/validate", "POST", map[string]string{"email": randomUser.Email, "validationCode": ""})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Verification code cannot be empty", response["message"])

	// -------------------------
	// 4. Test with incorrect validationCode
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/user/validate", "POST", map[string]string{"email": randomUser.Email, "validationCode": strconv.Itoa(codeNumber + 1)})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusUnauthorized, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Code was incorrect or time has expired", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}

// TestPasswordResetCodeSuccess Test the success case for /user/password/code endpoint
func TestPasswordResetCodeSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Verify the user directly on the database
	_, err := models.UserCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "isVerified", Value: true}}}})
	c.NoError(err)

	// Make the request and get the JSON response
	router.POST("/user/password/code", HandleResetPasswordCodeRequest)
	w, req := tests.SetupPayloadedRequest("/user/password/code", "POST", map[string]string{"email": randomUser.Email})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the response
	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal("New Code, to reset password, created and sended", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}

// TestPasswordResetCodeBadRequest Test the Bad Request cases for /user/password/code endpoint
func TestPasswordResetCodeBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	router := tests.SetupGinRouter()
	ctx := context.Background()
	defer ctx.Done()

	// -------------------------
	// 1. Test without JSON payload
	// -------------------------
	router.POST("/user/password/code", HandleResetPasswordCodeRequest)
	w, req := tests.SetupPayloadedRequest("/user/password/code", "POST", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Bad request", response["message"])

	// -------------------------
	// 2. Test with empty email
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/user/password/code", "POST", map[string]string{"email": ""})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Email cannot be empty", response["message"])

	// -------------------------
	// 3. Test with non-existent email
	// -------------------------
	w, req = tests.SetupPayloadedRequest("/user/password/code", "POST", map[string]string{"email": "645031e5-14da-45c8-abb7-714ded7d1ad9@gmail.com"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusNotFound, w.Code)
	c.Equal(true, response["error"])
	c.Equal("This Email has not been registered", response["message"])
}

// TestPasswordResetSuccess Test the success case for /user/password/reset endpoint
func TestPasswordResetSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Verify the user directly on the database
	_, err := models.UserCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "isVerified", Value: true}}}})
	c.NoError(err)

	// Send a request to get a new password reset code
	router.POST("/user/password/code", HandleResetPasswordCodeRequest)
	w, req := tests.SetupPayloadedRequest("/user/password/code", "POST", map[string]string{"email": randomUser.Email})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(http.StatusOK, w.Code)

	// Get the code from the database
	var passwordResetCode interfaces.AuthenticationCode
	err = models.AuthenticationCodesCollection.FindOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}, {Key: "type", Value: "RESET_PASSWORD"}}).Decode(&passwordResetCode)
	c.NoError(err)

	// -------------------------
	// 1. Test to reset the password
	router.PUT("/user/password/reset", HandleResetPassword)
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": randomUser.Email, "resetPassCode": passwordResetCode.Code, "password": "NewPassword2023*/"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Password has been changed successfully", response["message"])

	// -------------------------
	// 2. Test to log in with the new password
	router.POST("/session/login", HandleLogIn)
	w, req = tests.SetupPayloadedRequest("/session/login", "POST", map[string]string{"email": randomUser.Email, "password": "NewPassword2023*/"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the response
	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Successfully logged in", response["message"])

	// -------------------------
	// 3. Test to log in with the old password
	w, req = tests.SetupPayloadedRequest("/session/login", "POST", map[string]string{"email": randomUser.Email, "password": randomUser.Password})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	// Check the response
	c.Equal(http.StatusUnauthorized, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Wrong Email/Password", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}

// TestPasswordResetBadRequest Test the bad request cases for /user/password/reset endpoint
func TestPasswordResetBadRequest(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	randomUser := tests.GenerateRandomUser()
	router := tests.SetupGinRouter()
	tests.InsertUser(randomUser, router, HandleSignUp)

	// Verify the user directly on the database
	_, err := models.UserCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "isVerified", Value: true}}}})
	c.NoError(err)

	// Send a request to get a new password reset code
	router.POST("/user/password/code", HandleResetPasswordCodeRequest)
	w, req := tests.SetupPayloadedRequest("/user/password/code", "POST", map[string]string{"email": randomUser.Email})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(http.StatusOK, w.Code)

	// Get the code from the database
	var passwordResetCode interfaces.AuthenticationCode
	err = models.AuthenticationCodesCollection.FindOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}, {Key: "type", Value: "RESET_PASSWORD"}}).Decode(&passwordResetCode)
	c.NoError(err)

	// -------------------------
	// 1. Test to reset the password with nil payload
	router.PUT("/user/password/reset", HandleResetPassword)
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", nil)
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Bad request", response["message"])

	// -------------------------
	// 2. Test to reset the password with empty email
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": "", "resetPassCode": passwordResetCode.Code, "password": "NewPassword2023*/"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Email cannot be empty", response["message"])

	// -------------------------
	// 3. Test to reset the password with empty resetPassCode
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": randomUser.Email, "resetPassCode": "", "password": "NewPassword2023*/"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Verification code cannot be empty", response["message"])

	// -------------------------
	// 4. Test to reset the password with empty password
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": randomUser.Email, "resetPassCode": passwordResetCode.Code, "password": ""})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Password cannot be empty", response["message"])

	// -------------------------
	// 5. Test to reset the password with short password
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": randomUser.Email, "resetPassCode": passwordResetCode.Code, "password": "123"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Password must be at least 8 characters long", response["message"])

	// -------------------------
	// 6. Test to reset the password with invalid password (No uppercase)
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": randomUser.Email, "resetPassCode": passwordResetCode.Code, "password": "newpassword2023*/"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Password must have at least one upper case character", response["message"])

	// -------------------------
	// 7. Test to reset the password with invalid password (No lowercase)
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": randomUser.Email, "resetPassCode": passwordResetCode.Code, "password": "NEWPASSWORD2023*/"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Password must have at least one lower case character", response["message"])

	// -------------------------
	// 8. Test to reset the password with invalid password (No number)
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": randomUser.Email, "resetPassCode": passwordResetCode.Code, "password": "NewPassword*/"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Password must have at least one numeric character", response["message"])

	// -------------------------
	// 9. Test to reset the password with invalid password (No special character)
	w, req = tests.SetupPayloadedRequest("/user/password/reset", "PUT", map[string]string{"email": randomUser.Email, "resetPassCode": passwordResetCode.Code, "password": "NewPassword2023"})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Password must have at least one special character", response["message"])

	// Remove the user
	err = tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}

// TestGetLoomiesSuccess Test the success case for /user/loomies endpoint
func TestGetLoomiesSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	router := tests.SetupGinRouter()
	randomUser, loginResponse := loginWithRandomUser()
	token := loginResponse["accessToken"]

	// -------------------------
	// 1. Test with no loomies
	// -------------------------
	router.GET("/user/loomies", middlewares.MustProvideAccessToken(), HandleGetLoomies)
	w, req := tests.SetupGetRequest("/user/loomies", tests.CustomHeader{Name: "Access-Token", Value: token})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal(0, len(response["loomies"].([]interface{})))

	// -------------------------
	// 2. Test with some loomies
	// -------------------------
	// Get the first 6 loomies from the caught_loomies collection
	var loomies []interfaces.CaughtLoomie
	var loomiesIds []primitive.ObjectID
	cursor, _ := models.CaughtLoomiesCollection.Find(ctx, bson.D{}, options.Find().SetLimit(6).SetSort(bson.D{{Key: "_id", Value: -1}}))
	cursor.All(ctx, &loomies)
	c.Equal(6, len(loomies))

	// Get the ids of the loomies
	for _, loomie := range loomies {
		loomiesIds = append(loomiesIds, loomie.Id)
	}

	// Insert the loomies into the user's loomies
	models.UserCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "loomies", Value: loomiesIds}}}})

	// Make the request and get the JSON response
	w, req = tests.SetupGetRequest("/user/loomies", tests.CustomHeader{Name: "Access-Token", Value: token})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal(6, len(response["loomies"].([]interface{})))

	// Check the loomies fields
	for _, loomie := range response["loomies"].([]interface{}) {
		loomie := loomie.(map[string]interface{})
		c.NotEmpty(loomie["_id"])
		c.NotEmpty(loomie["serial"])
		c.NotEmpty(loomie["name"])
		c.NotEmpty(loomie["rarity"])
		c.NotEmpty(loomie["hp"])
		c.NotEmpty(loomie["attack"])
		c.NotEmpty(loomie["defense"])
		c.Contains(loomie, "is_busy")
		// Loomies cannot have less than 1 type
		c.NotEmpty(loomie["types"])
		c.GreaterOrEqual(len(loomie["types"].([]interface{})), 1)
		// Default level is 1
		c.NotEmpty(loomie["level"])
		c.GreaterOrEqual(int(loomie["level"].(float64)), 1)
		// Default experience is 0
		c.GreaterOrEqual(int(loomie["experience"].(float64)), 0)
	}

	// Remove the user
	err := tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}

// TestGetLoomieTeamSuccess Test the success case for /user/loomie-team endpoint
func TestGetLoomieTeamSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Create a random user
	router := tests.SetupGinRouter()
	randomUser, loginResponse := loginWithRandomUser()
	token := loginResponse["accessToken"]

	// -------------------------
	// 1. Test with no loomies in the team
	// -------------------------
	router.GET("/user/loomie-team", middlewares.MustProvideAccessToken(), HandleGetLoomieTeam)
	w, req := tests.SetupGetRequest("/user/loomie-team", tests.CustomHeader{Name: "Access-Token", Value: token})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal(0, len(response["team"].([]interface{})))

	// -------------------------
	// 2. Test with some loomies in the team
	// -------------------------
	// Get the first 6 loomies from the caught_loomies collection
	var loomies []interfaces.CaughtLoomie
	var loomiesIds []primitive.ObjectID

	// Select the last 6 loomies
	cursor, _ := models.CaughtLoomiesCollection.Find(ctx, bson.D{}, options.Find().SetLimit(6).SetSort(bson.D{{Key: "_id", Value: -1}}))
	cursor.All(ctx, &loomies)
	c.Equal(6, len(loomies))

	// Get the ids of the loomies
	for _, loomie := range loomies {
		loomiesIds = append(loomiesIds, loomie.Id)
	}

	// Insert the loomies into the user's loomies
	models.UserCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "loomies", Value: loomiesIds}}}})
	models.UserCollection.UpdateOne(ctx, bson.D{{Key: "email", Value: randomUser.Email}}, bson.D{{Key: "$set", Value: bson.D{{Key: "loomie_team", Value: loomiesIds}}}})

	// Set the user as the loomies in the array owner
	models.CaughtLoomiesCollection.UpdateMany(ctx, bson.D{
		{Key: "_id", Value: bson.D{{Key: "$in", Value: loomiesIds}}},
	}, bson.D{
		{Key: "$set", Value: bson.D{{Key: "owner", Value: randomUser.Id}}},
	})

	// Set the user as the loomies in the array owner
	models.CaughtLoomiesCollection.UpdateMany(ctx, bson.D{
		{Key: "_id", Value: bson.D{{Key: "$in", Value: loomiesIds}}},
	}, bson.D{
		{Key: "$set", Value: bson.D{{Key: "owner", Value: randomUser.Id}}},
	})

	// Make the request and get the JSON response
	w, req = tests.SetupGetRequest("/user/loomie-team", tests.CustomHeader{Name: "Access-Token", Value: token})
	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)

	c.Equal(http.StatusOK, w.Code)
	c.Equal(false, response["error"])
	c.Equal(6, len(response["team"].([]interface{})))

	// Check the loomies fields
	for _, loomie := range response["team"].([]interface{}) {
		c.NotEmpty(loomie)
		loomie := loomie.(map[string]interface{})
		c.NotEmpty(loomie["_id"])
		c.NotEmpty(loomie["serial"])
		c.NotEmpty(loomie["name"])
		c.NotEmpty(loomie["rarity"])
		c.NotEmpty(loomie["hp"])
		c.NotEmpty(loomie["attack"])
		c.NotEmpty(loomie["defense"])
		c.Contains(loomie, "is_busy")
		c.NotEmpty(loomie["types"])
		c.GreaterOrEqual(len(loomie["types"].([]interface{})), 1)
		c.NotEmpty(loomie["level"])
		c.GreaterOrEqual(int(loomie["level"].(float64)), 1)
		c.GreaterOrEqual(int(loomie["experience"].(float64)), 0)
	}

	// Remove the user
	err := tests.DeleteUser(randomUser.Email)
	c.NoError(err)
}
