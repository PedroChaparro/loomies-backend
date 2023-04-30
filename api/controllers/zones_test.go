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
	router.GET("/gyms/near", middlewares.MustProvideAccessToken(), HandleNearGyms)

	// ---- ---- ---- ----
	// Test 1: Test with an empty payload
	// ---- ---- ---- ----
	w, req := tests.SetupGetRequest("/gyms/near", tests.CustomHeader{
		Name:  "Access-Token",
		Value: loginResponse["accessToken"],
	})

	router.ServeHTTP(w, req)
	json.Unmarshal([]byte(w.Body.String()), &response)
	c.Equal(400, w.Code)
	c.Equal(true, response["error"])
	c.Equal("JSON payload is invalid or missing", response["message"])

	// Remove the user
	err := tests.DeleteUser(randomUser.Email, randomUser.Id)
	c.NoError(err)
}
