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

func InsertUser(data interfaces.User) error {
	// Set the current time as the "last time the user generated loomies"
	data.LastLoomieGenerationTime = time.Now().Unix()

	// Set the items and loomies as empty arrays
	data.Items = []interfaces.ItemsArray{}
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
