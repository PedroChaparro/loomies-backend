package models

import (
	"context"
	"fmt"
	"time"
	"unicode"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var Collection *mongo.Collection = configuration.ConnectToMongoCollection("users")

var LoomiesCollection *mongo.Collection = configuration.ConnectToMongoCollection("caught_loomies")

func InsertUser(data interfaces.User) error {
	// Set the current time as the "last time the user generated loomies"
	data.LastLoomieGenerationTime = time.Now().Unix()

	// Set the items and loomies as empty arrays
	data.Items = []interfaces.InventoryItem{}
	data.Loomies = []primitive.ObjectID{}

	//Insert User in database
	_, err := Collection.InsertOne(context.TODO(), data)

	return err
}

func ValidPassword(s string) error {
next:
	for name, classes := range map[string][]*unicode.RangeTable{
		"upper case": {unicode.Upper, unicode.Title},
		"lower case": {unicode.Lower},
		"numeric":    {unicode.Number, unicode.Digit},
		"special":    {unicode.Space, unicode.Symbol, unicode.Punct, unicode.Mark},
	} {
		for _, r := range s {
			if unicode.IsOneOf(classes, r) {
				continue next
			}
		}
		return fmt.Errorf("password must have at least one %s character", name)
	}
	return nil
}

// GetUserByEmail returns a user by its email and an error (if any)
func GetUserByEmail(email string) (interfaces.User, error) {
	var userE interfaces.User

	// Find the user with the given email (case insensitive)
	err := Collection.FindOne(
		context.TODO(),
		bson.M{"email": bson.M{"$regex": email, "$options": "i"}},
	).Decode(&userE)

	return userE, err
}

// CheckExistUsername returns an mongodb errror if the username already exists
func GetUserByUsername(Username string) (interfaces.User, error) {
	var userU interfaces.User

	//Find the user with the given username (case insensitive)
	err := Collection.FindOne(
		context.TODO(),
		bson.M{"username": bson.M{"$regex": Username, "$options": "i"}},
	).Decode(&userU)

	return userU, err
}

func GetUserByEmailAndVerifStatus(email string) (interfaces.User, error) {
	var userE interfaces.User
	err := Collection.FindOne(
		context.TODO(),
		bson.D{{Key: "email", Value: email}, {Key: "isVerified", Value: true}},
	).Decode(&userE)
	return userE, err
}

// GetUserById returns a user by its id
func GetUserById(id string) (interfaces.User, error) {
	var user interfaces.User

	mongoid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return user, err
	}

	err = Collection.FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: mongoid}},
	).Decode(&user)

	return user, err
}

// UpdateUserGenerationTimes updates the last time the user generated loomies and the current timeout
func UpdateUserGenerationTimes(userId string, lastGenerated int64, newTimeout int64) error {
	// Convert the string id to a mongo id
	mongoid, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		return err
	}

	// Update the user document
	_, err = Collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: mongoid}},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "lastLoomieGenerationTime", Value: lastGenerated},
				{Key: "currentLoomiesGenerationTimeout", Value: newTimeout},
			}},
		},
	)

	return err
}

func CheckCodeExistence(email string, code string) bool {

	var usercode interfaces.ValidationCode
	filter := bson.D{{Key: "email", Value: email}}
	err := Collection.FindOne(context.TODO(), filter).Decode(&usercode)
	if err != nil {
		fmt.Println(err)
	}
	// check time expire
	if !time.Now().Before(time.Unix(usercode.ValidationCodeExp, 0)) {
		// clean validationCode field
		filter := bson.D{{Key: "email", Value: usercode.Email}}
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "validationCode", Value: nil},
		},
		}}
		Collection.UpdateOne(context.TODO(), filter, update)
		return false
	}

	// verify code
	if code != usercode.ValidationCode {
		return false
	} else {
		// update status verified and delete one
		filter := bson.D{{Key: "email", Value: usercode.Email}}
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "isVerified", Value: true},
			{Key: "validationCode", Value: nil},
			{Key: "validationCodeExp", Value: nil},
		},
		}}
		Collection.UpdateOne(context.TODO(), filter, update)
		return true
	}
}

func UpdateCode(email string, validationCode string) error {
	// update code and expiration time
	filter := bson.D{{Key: "email", Value: email}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "validationCode", Value: validationCode},
		{Key: "validationCodeExp", Value: time.Now().Add(time.Minute * 15).Unix()},
	},
	}}
	_, err := Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// AddItemToUserInventory adds an item to the user's inventory
func AddItemToUserInventory(userId primitive.ObjectID, item interfaces.GymRewardItem) error {
	userInventoryItem := interfaces.InventoryItem{
		ItemCollection: item.RewardCollection,
		ItemId:         item.RewardId,
		ItemQuantity:   item.RewardQuantity,
	}

	// Update the user document
	_, err := Collection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: userId}},
		bson.D{
			{Key: "$push", Value: bson.D{
				{Key: "items", Value: userInventoryItem},
			}},
		},
	)

	return err
}

// AddItemsToUserInventory adds multiple items to the user's inventory
func AddItemsToUserInventory(userId primitive.ObjectID, items []interfaces.GymRewardItem) error {
	for _, item := range items {
		err := AddItemToUserInventory(userId, item)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetLoomiesByUser returns an array of loomies according with user
func GetLoomiesByUser(loomiesArray []primitive.ObjectID) ([]interfaces.UserLoomiesRes, error) {

	// array to keep loomies ids
	var loomiesIds []primitive.ObjectID = []primitive.ObjectID{}

	for _, element := range loomiesArray {
		loomiesIds = append(loomiesIds, element)
	}

	// Filter
	filter := bson.M{
		"_id": bson.M{
			"$in": loomiesIds,
		},
	}

	matchFilter := bson.M{"$match": filter}

	// Lookup loomies rarities collection
	lookupIntoRarity := bson.M{
		"$lookup": bson.M{
			"from":         "loomie_rarities",
			"localField":   "rarity",
			"foreignField": "_id",
			"as":           "rarity",
		},
	}

	// Lookup loomies types collection
	lookupIntoTypes := bson.M{
		"$lookup": bson.M{
			"from":         "loomie_types",
			"localField":   "types",
			"foreignField": "_id",
			"as":           "types",
		},
	}

	// Make the query
	cursor, err := LoomiesCollection.Aggregate(context.TODO(), []bson.M{matchFilter, lookupIntoRarity, lookupIntoTypes})

	if err != nil {
		return nil, err
	}

	var loomies []interfaces.UserLoomiesRes

	for cursor.Next(context.Background()) {
		var loomieAux interfaces.UserLoomiesResAux
		var loomie interfaces.UserLoomiesRes
		var types []string

		cursor.Decode(&loomieAux)
		cursor.Decode(&loomie)

		for _, v := range loomieAux.Types {
			types = append(types, v.Name)
		}

		loomie.Types = types

		loomies = append(loomies, loomie)
	}

	return loomies, err
}
