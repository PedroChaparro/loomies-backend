package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ZonesCollection *mongo.Collection = configuration.ConnectToMongoCollection("zones")

func GetZones() []interfaces.Zone {
	// Get zones from database
	cur, err := ZonesCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return []interfaces.Zone{}
	}

	// Create zones array
	var zones []interfaces.Zone
	for cur.Next(context.Background()) {
		var zone interfaces.Zone
		cur.Decode(&zone)
		zones = append(zones, zone)
	}

	return zones
}
