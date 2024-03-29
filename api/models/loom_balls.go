package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetLoomballsFromIds returns an array of loomballs from an array of loomballs ids
func GetLoomballsFromIds(ids []primitive.ObjectID) ([]interfaces.Loomball, error) {
	// If there are no ids, return an empty array to prevent errors
	if len(ids) == 0 {
		return []interfaces.Loomball{}, nil
	}

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
