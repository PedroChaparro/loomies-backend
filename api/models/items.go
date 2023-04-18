package models

import (
	"context"
	"fmt"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetItemFromUserInventory Returns the item from the user inventory
func GetItemFromUserInventory(userId primitive.ObjectID, itemId primitive.ObjectID) (interfaces.PopulatedInventoryItem, error) {
	// First we get the user who owns the item
	var user interfaces.User
	var item interfaces.PopulatedInventoryItem

	res := UserCollection.FindOne(context.TODO(), bson.M{
		"_id":           userId,
		"items.item_id": itemId,
	})

	if res.Err() == mongo.ErrNoDocuments {
		return interfaces.PopulatedInventoryItem{}, fmt.Errorf("USER_DOES_NOT_OWN_ITEM")
	}

	err := res.Decode(&user)

	if err != nil {
		return interfaces.PopulatedInventoryItem{}, err
	}

	// Get the user from the items collection
	res = ItemsCollection.FindOne(context.TODO(), bson.M{
		"_id": itemId,
	})

	if res.Err() == mongo.ErrNoDocuments {
		return interfaces.PopulatedInventoryItem{}, fmt.Errorf("ITEM_DOES_NOT_EXIST")
	}

	err = res.Decode(&item)

	if err != nil {
		return interfaces.PopulatedInventoryItem{}, err
	}

	// Complete the quantity field in the item
	for _, element := range user.Items {
		if element.ItemId == itemId {
			item.Quantity = element.ItemQuantity
		}
	}

	return item, nil
}

// GetItemById returns the items and loomballs from an array of items
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

		data = interfaces.UserItemsRes{Id: item.Id, Name: item.Name, Serial: item.Serial, Description: item.Description, Target: item.Target, Is_combat_item: item.IsCombatItem, Quantity: userItems[item.Id].ItemQuantity}
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

		data = interfaces.UserLoomballsRes{Id: loomball.Id, Name: loomball.Name, Serial: loomball.Serial, Quantity: userLoomballs[loomball.Id].ItemQuantity}
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

// GetItemFromUserInventory Returns the item from the user inventory
func GetItemFromUserInventory(userId primitive.ObjectID, itemId primitive.ObjectID, ignoreCombatItems bool) (interfaces.PopulatedInventoryItem, error) {
	// First we get the user who owns the item
	var user interfaces.User
	var item interfaces.PopulatedInventoryItem

	res := UserCollection.FindOne(context.TODO(), bson.M{
		"_id":           userId,
		"items.item_id": itemId,
	})

	if res.Err() == mongo.ErrNoDocuments {
		return interfaces.PopulatedInventoryItem{}, fmt.Errorf("USER_DOES_NOT_OWN_ITEM")
	}

	err := res.Decode(&user)
	if err != nil {
		return interfaces.PopulatedInventoryItem{}, err
	}

	// Get item user from the items collection
	if ignoreCombatItems {
		res = ItemsCollection.FindOne(context.TODO(), bson.M{
			"_id":            itemId,
			"is_combat_item": false,
		})
	} else {
		res = ItemsCollection.FindOne(context.TODO(), bson.M{
			"_id": itemId,
		})
	}

	if res.Err() == mongo.ErrNoDocuments {
		return interfaces.PopulatedInventoryItem{}, fmt.Errorf("ITEM_NOT_FOUND")
	}

	err = res.Decode(&item)
	if err != nil {
		return interfaces.PopulatedInventoryItem{}, err
	}

	// Complete the quantity field in the item
	for _, element := range user.Items {
		if element.ItemId == itemId {
			item.Quantity = element.ItemQuantity
		}
	}

	return item, nil
}
