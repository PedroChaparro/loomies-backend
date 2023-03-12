package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var Citems *mongo.Collection = configuration.ConnectToMongoCollection("items")
var ItemsCollection = configuration.ConnectToMongoCollection("items")

func GetItemById(itemsArray []interfaces.InventoryItem) ([]interfaces.PopulatedIventoryItem, error) {

	cursor, err := Citems.Find(context.TODO(), bson.D{})

	var user_items []interfaces.PopulatedIventoryItem

	for cursor.Next(context.TODO()) {
		var item interfaces.Item
		var data interfaces.PopulatedIventoryItem

		cursor.Decode(&item)

		for _, element := range itemsArray {

			if item.Id == element.Id {
				data = interfaces.PopulatedIventoryItem{Id: item.Id, Name: item.Name, Description: item.Description, Target: item.Target, Is_combat_item: item.Is_combat_item, Quantity: element.Quantity}
				user_items = append(user_items, data)
			}

		}

	}

	return user_items, err
}

// GetItemsFromIds returns an array of items from an array of items ids
func GetItemsFromIds(ids []primitive.ObjectID) ([]interfaces.Item, error) {
	// If there are no ids, return an empty array to prevent errors
	if len(ids) == 0 {
		return []interfaces.Item{}, nil
	}

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
