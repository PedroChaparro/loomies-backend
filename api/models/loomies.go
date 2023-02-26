package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var baseLoomiesCollection = configuration.ConnectToMongoCollection("base_loomies")
var wildLoomiesCollection = configuration.ConnectToMongoCollection("wild_loomies")

// GetBaseLoomies returns the base loomies
func GetBaseLoomies() ([]interfaces.BaseLoomiesWithPopulatedRarity, error) {
	baseLoomies := []interfaces.BaseLoomiesWithPopulatedRarity{}

	// Find all the base loomies and populate the rarity field
	lookupIntoRarities := bson.M{
		"$lookup": bson.M{
			"from":         "loomie_rarities",
			"localField":   "rarity",
			"foreignField": "_id",
			"as":           "populated_rarity",
		},
	}

	// Operations to perform on the populated rarity array
	aggProject := bson.M{
		"$project": bson.M{
			// Get the first element of the populated rarity array
			"populated_rarity": bson.M{
				"$arrayElemAt": []interface{}{"$populated_rarity", 0},
			},
			// Add the rest of the fields
			"name":         1,
			"serial":       1,
			"types":        1,
			"rarity":       1,
			"base_hp":      1,
			"base_attack":  1,
			"base_defense": 1,
		},
	}

	// Decode
	cursor, err := baseLoomiesCollection.Aggregate(context.TODO(), []bson.M{lookupIntoRarities, aggProject})

	if err != nil {
		return []interfaces.BaseLoomiesWithPopulatedRarity{}, err
	}

	cursor.All(context.Background(), &baseLoomies)

	return baseLoomies, err
}

// GetLoomiesFromZoneId returns the loomies that are in a zone
func GetLoomiesFromZoneId(id primitive.ObjectID) ([]interfaces.WildLoomie, error) {
	loomies := []interfaces.WildLoomie{}

	// Find all the loomies that are in the zone
	filter := bson.M{
		"zone_id": id,
	}

	cursor, err := wildLoomiesCollection.Find(context.Background(), filter)

	if err != nil {
		return []interfaces.WildLoomie{}, err
	}

	cursor.All(context.Background(), &loomies)

	return loomies, err
}

// InsertWildLoomie inserts a wild loomie into the database if the zone doesn't have the maximum amount of loomies
func InsertWildLoomie(loomie interfaces.WildLoomie) (interfaces.WildLoomie, bool) {
	// Get the zone coordinates
	coordX, coordY := utils.GetZoneCoordinatesFromGPS(interfaces.Coordinates{
		Latitude:  loomie.Latitude,
		Longitude: loomie.Longitude,
	})

	// Get the zone from the database
	zone, err := GetZoneFromCoordinates(coordX, coordY)

	if err != nil {
		return interfaces.WildLoomie{}, false
	}

	// Check if the zone has the maximum amount of loomies
	currentLoomies, err := GetLoomiesFromZoneId(zone.Id)
	// fmt.Println("Zone has", len(currentLoomies), "loomies")

	if err != nil || len(currentLoomies) >= configuration.GetMaxLoomiesPerZone() {
		// fmt.Println("Zone has the maximum amount of loomies")
		return interfaces.WildLoomie{}, false
	}

	// Insert the wild loomie into the database
	loomie.ZoneId = zone.Id
	result, err := wildLoomiesCollection.InsertOne(context.Background(), loomie)
	loomie.Id = result.InsertedID.(primitive.ObjectID)
	return loomie, err == nil
}
