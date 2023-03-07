package utils

import (
	"math"
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

func GetZoneCoordinatesFromGPS(coordinates interfaces.Coordinates) (int, int) {
	// initial zones calculations
	const initialLatitude = 6.9595
	const initialLongitude = -73.1696
	const sizeMinZone = 0.0035

	coordX := math.Floor((coordinates.Longitude - initialLongitude) / sizeMinZone)
	coordY := math.Floor((coordinates.Latitude - initialLatitude) / sizeMinZone)
	return int(coordX), int(coordY)
}

// get a code of 6 digits
func GetValidationCode() string {

	numbers := [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	var validationCode string = ""
	for i := 0; i < 6; i++ {
		validationCode += numbers[GetRandomInt(0, 9)]
	}

	return validationCode
}
