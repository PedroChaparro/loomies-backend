package models

import (
	"context"
	"fmt"
	"unicode"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var Collection *mongo.Collection = configuration.ConnectToMongoCollection("users")

func CheckExistEmail(email string) (interfaces.User, error) {
	var userE interfaces.User

	//Query in the database where the email
	err := Collection.FindOne(
		context.TODO(),
		bson.D{{Key: "email", Value: email}},
	).Decode(&userE)

	return userE, err

}

func CheckExistUsername(Username string) error {
	var userU interfaces.User

	//Query in the database where the username
	err := Collection.FindOne(
		context.TODO(),
		bson.D{{Key: "username", Value: Username}},
	).Decode(&userU)

	return err

}

func InsertUser(data interfaces.User) error {
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
