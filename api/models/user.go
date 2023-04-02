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

var UserCollection *mongo.Collection = configuration.ConnectToMongoCollection("users")

var LoomiesCollection *mongo.Collection = configuration.ConnectToMongoCollection("caught_loomies")

func InsertUser(data interfaces.User) error {
	// Set the current time as the "last time the user generated loomies"
	data.LastLoomieGenerationTime = time.Now().Unix()

	// Set the items and loomies as empty arrays
	data.Items = []interfaces.InventoryItem{}
	data.Loomies = []primitive.ObjectID{}

	//Insert User in database
	_, err := UserCollection.InsertOne(context.TODO(), data)

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

// update password when its reset
func UpdatePasword(email string, p string) error {
	filter := bson.D{{Key: "email", Value: email}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "password", Value: p},
	},
	}}
	_, err := UserCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// GetUserByEmail returns a user by its email and an error (if any)
func GetUserByEmail(email string) (interfaces.User, error) {
	var userE interfaces.User

	// Find the user with the given email (case insensitive)
	err := UserCollection.FindOne(
		context.TODO(),
		bson.M{"email": bson.M{"$regex": email, "$options": "i"}},
	).Decode(&userE)

	return userE, err
}

// CheckExistUsername returns an mongodb errror if the username already exists
func GetUserByUsername(Username string) (interfaces.User, error) {
	var userU interfaces.User

	//Find the user with the given username (case insensitive)
	err := UserCollection.FindOne(
		context.TODO(),
		bson.M{"username": bson.M{"$regex": Username, "$options": "i"}},
	).Decode(&userU)

	return userU, err
}

func GetUserByEmailAndVerifStatus(email string) (interfaces.User, error) {
	var userE interfaces.User
	err := UserCollection.FindOne(
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

	err = UserCollection.FindOne(
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
	_, err = UserCollection.UpdateOne(
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
	var codeDoc interfaces.AuthenticationCode
	filter := bson.D{{Key: "email", Value: email}, {Key: "type", Value: "ACCOUNT_VERIFICATION"}}
	err := AuthenticationCodesCollection.FindOne(context.TODO(), filter).Decode(&codeDoc)

	if err != nil {
		fmt.Println(err)
		return false
	}

	// check time expire
	if !time.Now().Before(time.Unix(codeDoc.ExpiresAt, 0)) {
		// If the code has expired, delete it
		AuthenticationCodesCollection.DeleteOne(context.TODO(), filter)
		return false
	}

	// Check if the code is correct
	if code != codeDoc.Code {
		return false
	} else {
		// Update the user docuement
		filter := bson.D{{Key: "email", Value: email}}

		update := bson.D{
			{
				Key: "$set", Value: bson.D{
					{Key: "isVerified", Value: true},
				},
			},
		}

		_, err := UserCollection.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			fmt.Println(err)
			return false
		}

		// Remove the code
		_, err = AuthenticationCodesCollection.DeleteOne(context.TODO(), filter)

		if err != nil {
			// Print the error but return true because at this point, the user
			// is verified
			fmt.Println(err)
		}

		return true
	}
}

func CheckResetPassCodeExistence(email string, code string) bool {
	var codeDoc interfaces.AuthenticationCode
	filter := bson.D{{Key: "email", Value: email}, {Key: "type", Value: "RESET_PASSWORD"}}
	err := AuthenticationCodesCollection.FindOne(context.TODO(), filter).Decode(&codeDoc)

	if err != nil {
		fmt.Println(err)
		return false
	}

	// If the code has expired, delete it and return false
	if !time.Now().Before(time.Unix(codeDoc.ExpiresAt, 0)) {
		AuthenticationCodesCollection.DeleteOne(context.TODO(), filter)
		return false
	}

	// Check the given code is equal
	if code != codeDoc.Code {
		return false
	} else {
		// If the code is correct, delete it and return true
		AuthenticationCodesCollection.DeleteOne(context.TODO(), filter)
		return true
	}
}

func UpdateCode(email string, validationCode string) error {
	// Insert the code in the codes collection
	_, err := AuthenticationCodesCollection.InsertOne(context.TODO(), interfaces.AuthenticationCode{
		Email:     email,
		Code:      validationCode,
		Type:      "ACCOUNT_VERIFICATION",
		ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
	})

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("Error while inserting the code in the database")
	}

	return nil
}

// update the fields of code to reset password and its expiration time in db
func UpdateResetPassCode(email string, resetPassCode string) error {
	// Remove older codes
	filter := bson.D{{Key: "email", Value: email}, {Key: "type", Value: "RESET_PASSWORD"}}
	_, err := AuthenticationCodesCollection.DeleteMany(context.TODO(), filter)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("Error deleting old reset password codes")
	}

	// Insert the new code
	_, err = AuthenticationCodesCollection.InsertOne(context.TODO(), interfaces.AuthenticationCode{
		Email:     email,
		Code:      resetPassCode,
		Type:      "RESET_PASSWORD",
		ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
	})

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("Error inserting new reset password code")
	}

	return nil
}

// AddItemToUserInventory adds an item to the user's inventory
func AddItemToUserInventory(userId primitive.ObjectID, item interfaces.GymRewardItem) error {
	userInventoryItem := interfaces.InventoryItem{
		ItemCollection: item.RewardCollection,
		ItemId:         item.RewardId,
		ItemQuantity:   item.RewardQuantity,
	}

	// Check if the item already exists in the user's inventory
	var user interfaces.User
	var alreadyExists = false

	err := UserCollection.FindOne(
		context.TODO(),
		bson.D{
			{Key: "_id", Value: userId},
		},
	).Decode(&user)

	if err != nil {
		return err
	}

	for _, inventoryItem := range user.Items {
		if inventoryItem.ItemId == item.RewardId {
			alreadyExists = true
			break
		}
	}

	if alreadyExists {
		// Update the user document to increment the item quantity
		_, err = UserCollection.UpdateOne(
			context.TODO(),
			bson.D{
				// Match tue user
				{Key: "_id", Value: userId},
				// Match the item inside the user's inventory
				{Key: "items", Value: bson.D{
					{Key: "$elemMatch", Value: bson.D{
						{Key: "item_id", Value: item.RewardId},
					},
					}},
				},
			}, bson.D{
				// Increment the item quantity
				{Key: "$inc", Value: bson.D{
					{Key: "items.$.item_quantity", Value: item.RewardQuantity},
				}},
			})
	} else {
		// Update the user document to add the item to the inventory
		_, err = UserCollection.UpdateOne(
			context.TODO(),
			bson.D{{Key: "_id", Value: userId}},
			bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "items", Value: userInventoryItem},
				}},
			},
		)
	}

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
func GetLoomiesByIds(loomiesArray []primitive.ObjectID) ([]interfaces.UserLoomiesRes, error) {

	// Filter
	filter := bson.M{
		"_id": bson.M{
			"$in": loomiesArray,
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

		for _, t := range loomieAux.Types {
			types = append(types, t.Name)
		}

		loomie.Rarity = loomieAux.Rarity[0].Name
		loomie.Types = types

		loomies = append(loomies, loomie)
	}

	return loomies, err
}
