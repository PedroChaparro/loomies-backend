package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var LoomballsCollection = configuration.ConnectToMongoCollection("loom_balls")

// GetLoomballsFromIds returns an array of loomballs from an array of loomballs ids
func GetLoomballsFromIds(ids []primitive.ObjectID) ([]interfaces.Loomball, error) {
	var loomballs []interfaces.Loomball
	cursor, err := LoomballsCollection.Find(context.Background(), bson.M{"_id": bson.M{"$in": ids}})

	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.Background(), &loomballs); err != nil {
		return nil, err
	}

	return loomballs, nil
}
