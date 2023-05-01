package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/PedroChaparro/loomies-backend/middlewares"
	"github.com/PedroChaparro/loomies-backend/tests"
	"github.com/stretchr/testify/require"
)

// TestNearGymsErrors Test the error responses for the `/gyms/near` endpoint
func TestNearGymsErrors(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/gyms/near", middlewares.MustProvideAccessToken(), HandleNearGyms)

	// ---- ---- ---- ----
	// Test 1: Test with an empty payload
	// ---- ---- ---- ----
	w, req := tests.SetupPayloadedRequest("/gyms/near", "POST", nil, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("JSON payload is invalid or missing", response["message"])

	// Remove the user
	err := tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}

// TestNearGymsSuccess Test the success responses for the `/gyms/near` endpoint
func TestNearGymsSuccess(t *testing.T) {
	var response map[string]interface{}
	c := require.New(t)
	ctx := context.Background()
	defer ctx.Done()

	// Login with a random user
	randomUser, loginResponse := loginWithRandomUser()

	// Setup the router
	router := tests.SetupGinRouter()
	router.POST("/gyms/near", middlewares.MustProvideAccessToken(), HandleNearGyms)

	// ---- ---- ---- ----
	// Test 1: Test the UPB coordinates
	// ---- ---- ---- ----
	w, req := tests.SetupPayloadedRequest("/gyms/near", "POST", map[string]interface{}{
		"latitude":  7.03825,
		"longitude": -73.07138,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(200, w.Code)
	c.Equal(false, response["error"])
	c.Equal("Gyms have been found in near areas", response["message"])
	c.GreaterOrEqual(len(response["nearGyms"].([]interface{})), 12)

	// ---- ---- ---- ----
	// Test 2: Test with far away coordinates
	// ---- ---- ---- ----
	w, req = tests.SetupPayloadedRequest("/gyms/near", "POST", map[string]interface{}{
		"latitude":  4.60169,
		"longitude": -74.07198,
	}, tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &response)
	c.Equal(404, w.Code)
	c.Equal(true, response["error"])
	c.Equal("Gyms Not Found", response["message"])
	c.Equal(0, len(response["nearGyms"].([]interface{})))

	// Remove the user
	err := tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}
