package models

import (
	"context"
	"fmt"
	"unicode"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var Collection *mongo.Collection = configuration.ConnectToMongoCollection("users")
var User interfaces.User

func CheckExistEmail(email string) error {

	err := Collection.FindOne(
		context.TODO(),
		bson.D{{"email", email}},
	).Decode(&User)

	return err

}

func CheckExistUser(user string) error {

	err := Collection.FindOne(
		context.TODO(),
		bson.D{{"user", user}},
	).Decode(&User)

	return err

}

func InsertUser(data map[string]interface{}) error {

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
