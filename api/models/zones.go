package models

import (
	"context"
	"fmt"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
)

// GetNearGyms Returns an array of gyms near the current coordinates
func GetNearGyms(currentLatitude float64, currentLongitude float64) (p []interfaces.NearGymsRes, e error) {
	mZonesCoord := utils.GetNearZonesCoordinates(interfaces.Coordinates{
		Latitude:  currentLatitude,
		Longitude: currentLongitude,
	})

	// filters for aggregation and lookup into gyms collection
	zonesFilter := bson.M{"coordinates": bson.M{"$in": mZonesCoord}}
	matchFilter := bson.M{"$match": zonesFilter}
	lookupIntoGyms := bson.M{
		"$lookup": bson.M{
			"from":         "gyms",
			"localField":   "gyms",
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

		for _, gym := range surroundZones.Gyms {
			// Parse the gym to NearGymsRes (to remove unnecessary fields)
			nearGymStruct := gym.ToNearGymsRes()
			gyms = append(gyms, *nearGymStruct)
		}
	}

	return gyms, err
}

// GetZoneFromCoordinates returns a zone from the given coordinates
func GetZoneFromCoordinates(coordX int, coordY int) (interfaces.Zone, error) {
	var zone interfaces.Zone
	zoneFilter := bson.M{"coordinates": fmt.Sprintf("%v,%v", coordX, coordY)}
	err := ZonesCollection.FindOne(context.Background(), zoneFilter).Decode(&zone)
	return zone, err
}
