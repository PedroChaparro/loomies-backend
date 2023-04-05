package utils

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
	"unicode"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
)

// CheckPasswordSchema checks if the given password is valid
func CheckPasswordSchema(s string) error {
next:
	for name, classes := range map[string][]*unicode.RangeTable{
		"upper case": {unicode.Upper, unicode.Title},
		"lower case": {unicode.Lower},
		"numeric":    {unicode.Number, unicode.Digit},
		"special":    {unicode.Space, unicode.Symbol, unicode.Punct, unicode.Mark},
	} {
		for _, r := range s {
			if unicode.IsOneOf(classes, r) {
				continue next
			}
		}
		return fmt.Errorf("Password must have at least one %s character", name)
	}
	return nil
}

// GetRandomInt returns a random integer between min and max (both included)
func GetRandomInt(min int, max int) int {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	return r.Intn(max-min) + min
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

// GetZoneCoordinatesFromGPS returns the (x, y) coordinates of the zone that contains the given coordinates
func GetZoneCoordinatesFromGPS(coordinates interfaces.Coordinates) (int, int) {
	// initial zones calculations
	const initialLatitude = 6.9595
	const initialLongitude = -73.1696
	const sizeMinZone = 0.0035

	coordX := math.Floor((coordinates.Longitude - initialLongitude) / sizeMinZone)
	coordY := math.Floor((coordinates.Latitude - initialLatitude) / sizeMinZone)
	return int(coordX), int(coordY)
}

// GetValidationCode returns a random 6 digit string
func GetValidationCode() string {
	numbers := [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	var validationCode string = ""
	for i := 0; i < 6; i++ {
		validationCode += numbers[GetRandomInt(0, 9)]
	}

	return validationCode
}

// IsNear returns true if the target coordinates are near the origin coordinates
func IsNear(target interfaces.Coordinates, origin interfaces.Coordinates) bool {
	zoneRadiusStr := configuration.GetEnvironmentVariable("GAME_ZONE_RADIUS")
	zoneRadius, _ := strconv.ParseFloat(zoneRadiusStr, 64)

	if math.Abs(target.Latitude-origin.Latitude) > zoneRadius || math.Abs(target.Longitude-origin.Longitude) > zoneRadius {
		return false
	}

	return true
}

// GetLoomiesExperience returns the experience needed to reach the given level
func GetRequiredExperience(level int) float64 {
	min, factor := configuration.GetLoomiesExperienceParameters()
	return math.Log10(float64(level))*factor + min
}

// GetLevelFromExperience returns the level of the given experience
func GetLevelFromExperience(experience float64) int {
	min, factor := configuration.GetLoomiesExperienceParameters()
	return int(math.Pow(10, (experience-min)/factor))
}

// FixeFloat Returns the given float with the given number of decimals
func FixeFloat(float float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return float64(math.Round(float*pow)) / pow
}
