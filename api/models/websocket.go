package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/configuration"
	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var CombatCollection = configuration.ConnectToMongoCollection("combats")

// InitializeCombat creates a new combat in the database from the
// token claims
func InitializeCombat(claims interfaces.WsTokenClaims) error {
	// Parse the user and gym ids from the claims
	mongoUserId, _ := primitive.ObjectIDFromHex(claims.UserID)
	mongoGymId, _ := primitive.ObjectIDFromHex(claims.GymID)

	combat := interfaces.Combat{
		PlayerId: mongoUserId,
		GymId:    mongoGymId,
	}

	// Insert to the combats collection
	_, err := CombatCollection.InsertOne(context.Background(), combat)
	return err
}
