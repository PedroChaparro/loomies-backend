package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ItemsCollection = configuration.ConnectToMongoCollection("items")

// GetItemsFromIds returns an array of items from an array of items ids
func GetItemsFromIds(ids []primitive.ObjectID) ([]interfaces.Item, error) {
	var itemsE []interfaces.Item
	cursor, err := ItemsCollection.Find(context.Background(), bson.M{"_id": bson.M{"$in": ids}})

	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.Background(), &itemsE); err != nil {
		return nil, err
	}

	return itemsE, nil
}
