package utils

import (
	"math/rand"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
)

// GetRandomInt returns a random integer between min and max (both included)
func GetRandomInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// GetRandomFloat returns a random float64 between min and max (both included)
func GetRandomFloat(min float64, max float64) float64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Float64()*(max-min) + min
}

// GetRandomCoordinatesNear returns a random coordinates near the given coordinates
func GetRandomCoordinatesNear(coordinates interfaces.Coordinates) interfaces.Coordinates {
	radius := configuration.GetLoomiesGenerationRadius()
	latitude := GetRandomFloat(coordinates.Latitude-radius, coordinates.Latitude+radius)
	longitude := GetRandomFloat(coordinates.Longitude-radius, coordinates.Longitude+radius)

	return interfaces.Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	}
}
