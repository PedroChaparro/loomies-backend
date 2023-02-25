package models

import (
	"context"
	"fmt"
	"math"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ZonesCollection *mongo.Collection = configuration.ConnectToMongoCollection("zones")

func GetNearGyms(currentLatitude float64, currentLongitude float64) (p []interfaces.Gym, e error) {

	// initial zones calculations
	const initialLatitude = 6.9595
	const initialLongitude = -73.1696
	const sizeMinZone = 0.0035

	coordX := math.Floor((currentLongitude - initialLongitude) / sizeMinZone)
	coordY := math.Floor((currentLatitude - (initialLatitude)) / sizeMinZone)

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

	var gyms []interfaces.Gym

	for cursor.Next(context.Background()) {
		var surroundZones interfaces.ZoneWithGyms
		cursor.Decode(&surroundZones)
		// some zones don't have gyms
		if len(surroundZones.Gyms) != 0 {
			gyms = append(gyms, surroundZones.Gyms[0])
		}
	}

	return gyms, err

}
