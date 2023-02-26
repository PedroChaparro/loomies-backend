package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
)

var baseLoomiesCollection = configuration.ConnectToMongoCollection("base_loomies")

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
