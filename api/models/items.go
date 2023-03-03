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

func GetItemById(id primitive.ObjectID) (interfaces.Items, error) {
	var item interfaces.Items

	err := Citems.FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: id}},
	).Decode(&item)

	return item, err
}
