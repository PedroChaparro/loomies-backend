package models

import (
	"context"
	"fmt"
	"math"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
)

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

// GetNearGyms Returns an array of gyms near the current coordinates
func GetNearGyms(currentLatitude float64, currentLongitude float64) (p []interfaces.NearGymsRes, e error) {
	coordX, coordY := GetZoneCoordinatesFromGPS(interfaces.Coordinates{
		Latitude:  currentLatitude,
		Longitude: currentLongitude,
	})

	var mZonesCoord []string
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX-1, coordY+1)) // Box Top Left
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX, coordY+1))   // Box Top - North
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX+1, coordY+1)) // Box Top Right
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX-1, coordY))   // Box Left
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX, coordY))     // current zone box
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX+1, coordY))   // Box Right
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX-1, coordY-1)) // Box Bottom Left
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX, coordY-1))   // Box Bottom - South
	mZonesCoord = append(mZonesCoord, fmt.Sprintf("%v,%v", coordX+1, coordY-1)) // Box Bottom Right

	// filters for aggregation and lookup into gyms collection
	zonesFilter := bson.M{"coordinates": bson.M{"$in": mZonesCoord}}
	matchFilter := bson.M{"$match": zonesFilter}
	lookupIntoGyms := bson.M{
		"$lookup": bson.M{
			"from":         "gyms",
			"localField":   "gym",
			"foreignField": "_id",
			"as":           "gyms",
		},
	}

	// make the aggregation and get the gym field
	cursor, err := ZonesCollection.Aggregate(context.TODO(), []bson.M{matchFilter, lookupIntoGyms})

	var gyms []interfaces.NearGymsRes

	for cursor.Next(context.Background()) {
		var surroundZones interfaces.ZoneWithGyms
		cursor.Decode(&surroundZones)

		// some zones don't have gyms
		if len(surroundZones.Gyms) != 0 {
			// Parse the gym to NearGymsRes (to remove unnecessary fields)
			nearGymStruct := surroundZones.Gyms[0].ToNearGymsRes()
			gyms = append(gyms, *nearGymStruct)
		}
	}

	fmt.Println(err)
	return gyms, err
}

// GetZoneFromCoordinates returns a zone from the given coordinates
func GetZoneFromCoordinates(coordX int, coordY int) (interfaces.Zone, error) {
	var zone interfaces.Zone
	zoneFilter := bson.M{"coordinates": fmt.Sprintf("%v,%v", coordX, coordY)}
	err := ZonesCollection.FindOne(context.Background(), zoneFilter).Decode(&zone)
	return zone, err
}
