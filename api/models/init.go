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

func GetCurrentZone(latitude float32, longitude float32) (p interfaces.Zone, e error) {

	filter := bson.D{
		{"$and",
			bson.A{
				bson.D{{"rightFrontier", bson.D{{"$gte", latitude}}}},
				bson.D{{"leftFrontier", bson.D{{"$lte", latitude}}}},
				bson.D{{"bottomFrontier", bson.D{{"$lte", longitude}}}},
				bson.D{{"topFrontier", bson.D{{"$gte", longitude}}}},
			},
		},
	}
	// find current location
	var results interfaces.Zone
	err := ZonesCollection.FindOne(
		context.Background(),
		filter).Decode(&results)

	if err != nil {
		return interfaces.Zone{}, err
	}

	/* if err = currLocation.All(context.TODO(), &results); err != nil {
		panic(err)
	} */

	/* filterNearZones := bson.D{
		{"$and",
			bson.A{
				bson.D{{"rightFrontier", bson.D{{"$lte", (latitude + 0.0035)}}}},
				bson.D{{"leftFrontier", bson.D{{"$gte", (latitude + 0.0035)}}}},
				bson.D{{"bottomFrontier", bson.D{{"$gte", (longitude + 0.0035)}}}},
				bson.D{{"topFrontier", bson.D{{"$lte", (longitude + 0.0035)}}}},
			},
		},
	} */

	/* for _, result := range results {
		res, _ := json.Marshal(result)

		fmt.Println(result)
		fmt.Println(string(res))
	} */

	return results, err

}
