package models

import (
	"context"
	"fmt"
	"time"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var zonesCollection = configuration.ConnectToMongoCollection("zones")
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
	// fmt.Println("Getting zone from coordinates", coordX, coordY)
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
	loomie.GeneratedAt = time.Now().Unix()
	result, err := wildLoomiesCollection.InsertOne(context.Background(), loomie)

	if err != nil {
		return interfaces.WildLoomie{}, false
	}

	// Update the loomies array in the zone
	_, err = zonesCollection.UpdateOne(context.Background(), bson.M{"_id": zone.Id}, bson.M{"$push": bson.M{"loomies": result.InsertedID}})
	loomie.Id = result.InsertedID.(primitive.ObjectID)

	return loomie, err == nil
}

// GetNearWildLoomies returns the wild loomies that are near the coordinates
func GetNearWildLoomies(coordinates interfaces.Coordinates) ([]interfaces.WildLoomie, error) {
	candidateLoomies := []interfaces.WildLoomie{}
	loomies := []interfaces.WildLoomie{}

	// Get the zone coordinates
	coordX, coordY := utils.GetZoneCoordinatesFromGPS(coordinates)

	// Get the zones that are near the current zone
	var nearZonesCoordinates []string
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX-1, coordY+1)) // Box Top Left
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX, coordY+1))   // Box Top - North
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX+1, coordY+1)) // Box Top Right
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX-1, coordY))   // Box Left
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX, coordY))     // current zone box
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX+1, coordY))   // Box Right
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX-1, coordY-1)) // Box Bottom Left
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX, coordY-1))   // Box Bottom - South
	nearZonesCoordinates = append(nearZonesCoordinates, fmt.Sprintf("%v,%v", coordX+1, coordY-1)) // Box Bottom Right

	// Filter
	zonesFilter := bson.M{"coordinates": bson.M{"$in": nearZonesCoordinates}}
	matchFilter := bson.M{"$match": zonesFilter}

	// Aggregation to populate zone's loomies
	lookupIntoLoomies := bson.M{
		"$lookup": bson.M{
			"from":         "wild_loomies",
			"localField":   "loomies",
			"foreignField": "_id",
			"as":           "populated_loomies",
		},
	}

	// Make the query
	cursor, err := zonesCollection.Aggregate(context.Background(), []bson.M{matchFilter, lookupIntoLoomies})

	if err != nil {
		return []interfaces.WildLoomie{}, err
	}

	for cursor.Next(context.Background()) {
		var zone interfaces.ZoneWithPopulatedLoomies
		cursor.Decode(&zone)
		candidateLoomies = append(candidateLoomies, zone.PopulatedLoomies...)
	}

	loomieTTL := configuration.GetWildLoomiesTTL()
	currentTime := time.Now()

	// Keep only the loomies that are not expired
	for _, loomie := range candidateLoomies {
		loomieDeadline := time.Unix(loomie.GeneratedAt, 0).Add(time.Minute * time.Duration(loomieTTL))

		if currentTime.Before(loomieDeadline) {
			loomies = append(loomies, loomie)
		}
	}

	return loomies, nil
}

// ValidateLoomieExists Check if the loomie exists by the id
func ValidateLoomieExists(loomie_id string) error {
	id, err := primitive.ObjectIDFromHex(loomie_id)
	var loomie interfaces.WildLoomie

	if err != nil {
		return err
	}

	err = wildLoomiesCollection.FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: id}},
	).Decode(&loomie)

	return err
}
