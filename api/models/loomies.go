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
	cursor, err := BaseLoomiesCollection.Aggregate(context.TODO(), []bson.M{lookupIntoRarities, aggProject})

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

	cursor, err := WildLoomiesCollection.Find(context.Background(), filter)

	if err != nil {
		return []interfaces.WildLoomie{}, err
	}

	cursor.All(context.Background(), &loomies)

	return loomies, err
}

// InsertWildLoomie inserts a wild loomie into the database if the zone doesn't have the maximum amount of loomies
func InsertWildLoomie(loomie interfaces.WildLoomie) (interfaces.WildLoomie, bool) {
	// Get the zone coordinates
	coordX, coordY := GetZoneCoordinatesFromGPS(interfaces.Coordinates{
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
	result, err := WildLoomiesCollection.InsertOne(context.Background(), loomie)

	if err != nil {
		return interfaces.WildLoomie{}, false
	}

	// Update the loomies array in the zone
	_, err = ZonesCollection.UpdateOne(context.Background(), bson.M{"_id": zone.Id}, bson.M{"$push": bson.M{"loomies": result.InsertedID}})
	loomie.Id = result.InsertedID.(primitive.ObjectID)

	return loomie, err == nil
}

// GetNearWildLoomies returns the wild loomies that are near the coordinates
func GetNearWildLoomies(coordinates interfaces.Coordinates) ([]interfaces.WildLoomie, error) {
	candidateLoomies := []interfaces.WildLoomie{}
	loomies := []interfaces.WildLoomie{}

	// Get the zone coordinates
	coordX, coordY := GetZoneCoordinatesFromGPS(coordinates)

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
	cursor, err := ZonesCollection.Aggregate(context.Background(), []bson.M{matchFilter, lookupIntoLoomies})

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

// GetWildLoomieById Returns the wild loomie with the given id
func GetWildLoomieById(loomie_id string) (interfaces.WildLoomie, error) {
	id, err := primitive.ObjectIDFromHex(loomie_id)
	var loomie interfaces.WildLoomie

	if err != nil {
		return loomie, err
	}

	err = WildLoomiesCollection.FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: id}},
	).Decode(&loomie)

	return loomie, err
}

// InsertInCaughtLoomies Insert the loomie in the caught loomies collection
func InsertInCaughtLoomies(caught_loomie interfaces.CaughtLoomie) (primitive.ObjectID, error) {
	result, err := CaughtLoomiesCollection.InsertOne(context.TODO(), caught_loomie)

	if err != nil {
		return primitive.NilObjectID, err
	}

	id, _ := result.InsertedID.(primitive.ObjectID)

	return id, err
}

// WasSuccessfulCapture Check if the loomie was successful capture (Calculate the chance of success)
func WasSuccessfulCapture(loomie interfaces.WildLoomie, ball interfaces.Loomball) bool {
	chance := 0
	capture := utils.GetRandomInt(0, 100)

	if loomie.Level >= ball.DecayUntil {
		chance = int(ball.MinimumProbability * 100)
	} else if loomie.Level <= int(ball.EffectiveUntil) {
		chance = 100
	} else {
		chance = -((100-int(ball.MinimumProbability*100))/(ball.DecayUntil-ball.EffectiveUntil))*(loomie.Level-ball.EffectiveUntil) + 100
	}

	if capture <= chance {
		return true
	}
	return false
}

// CheckIfUserInArrayOfWildLoomie check if user id alrady exists in array CapturedBy from wild loomie
func CheckIfUserInArrayOfWildLoomie(loomie interfaces.WildLoomie, user interfaces.User) bool {
	for _, element := range loomie.CapturedBy {
		if element == user.Id {
			return false
		}
	}

	return true
}

// InsertUserInArrayOfWildLoomie insert user id in array CapturedBy from wild loomie
func InsertUserInArrayOfWildLoomie(loomie interfaces.WildLoomie, user interfaces.User) error {
	filter := bson.D{{Key: "_id", Value: loomie.Id}}
	update := bson.D{{Key: "$push", Value: bson.D{
		{Key: "captured_by", Value: user.Id},
	},
	}}
	_, err := WildLoomiesCollection.UpdateOne(context.TODO(), filter, update)

	return err
}

// GetLoomieTypeDetails Returns the details of a loomie type
func GetLoomieTypeDetailsByName(typeName string) (interfaces.PopulatedLoomieType, error) {
	var loomieTypeAuxiliar interfaces.PopulatedLoomieTypeAuxiliar
	var loomieType interfaces.PopulatedLoomieType

	// Querty the database
	lookupIntoTypes := bson.M{
		"$lookup": bson.M{
			"from":         "loomie_types",
			"localField":   "strong_against",
			"foreignField": "_id",
			"as":           "strong_against",
		},
	}

	cursor, err := LoomieTypesCollection.Aggregate(context.Background(), []bson.M{
		{"$match": bson.M{"name": typeName}},
		lookupIntoTypes,
	})

	if err != nil {
		return interfaces.PopulatedLoomieType{}, err
	}

	// Parse the result
	if cursor.Next(context.Background()) {
		err = cursor.Decode(&loomieTypeAuxiliar)
	}

	// Convert auxiliar to the final type
	loomieType.Id = loomieTypeAuxiliar.Id
	loomieType.Name = loomieTypeAuxiliar.Name

	for _, strongAgainst := range loomieTypeAuxiliar.StrongAgainst {
		loomieType.StrongAgainst = append(loomieType.StrongAgainst, strongAgainst.Name)
	}

	return loomieType, err
}
