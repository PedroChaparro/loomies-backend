package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetItemById(itemsArray []interfaces.InventoryItem) ([]interfaces.UserItemsRes, []interfaces.UserLoomballsRes, error) {
	// Create the maps to store the items to access them faster
	userItems := make(map[primitive.ObjectID]interfaces.InventoryItem)
	userLoomballs := make(map[primitive.ObjectID]interfaces.InventoryItem)

	// Save the ids to easily query the database
	var itemsIds []primitive.ObjectID = []primitive.ObjectID{}
	var loomballsIds []primitive.ObjectID = []primitive.ObjectID{}

	for _, element := range itemsArray {
		if element.ItemCollection == "items" {
			userItems[element.ItemId] = element
			itemsIds = append(itemsIds, element.ItemId)
		} else {
			userLoomballs[element.ItemId] = element
			loomballsIds = append(loomballsIds, element.ItemId)
		}
	}

	// Get the items from the database
	var items []interfaces.UserItemsRes

	cursor, err := ItemsCollection.Find(context.TODO(), bson.M{
		"_id": bson.M{
			"$in": itemsIds,
		},
	})

	if err != nil {
		return nil, nil, err
	}

	for cursor.Next(context.TODO()) {
		var item interfaces.Item
		var data interfaces.UserItemsRes
		cursor.Decode(&item)

		data = interfaces.UserItemsRes{Id: item.Id, Name: item.Name, Description: item.Description, Target: item.Target, Is_combat_item: item.IsCombatItem, Quantity: userItems[item.Id].ItemQuantity}
		items = append(items, data)
	}

	// Get the loomballs from the database
	var loomballs []interfaces.UserLoomballsRes

	cursor, err = LoomballsCollection.Find(context.TODO(), bson.M{
		"_id": bson.M{
			"$in": loomballsIds,
		},
	})

	if err != nil {
		return nil, nil, err
	}

	for cursor.Next(context.TODO()) {
		var loomball interfaces.Loomball
		var data interfaces.UserLoomballsRes
		cursor.Decode(&loomball)

		data = interfaces.UserLoomballsRes{Id: loomball.Id, Name: loomball.Name, Quantity: userLoomballs[loomball.Id].ItemQuantity}
		loomballs = append(loomballs, data)
	}

	return items, loomballs, err
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
