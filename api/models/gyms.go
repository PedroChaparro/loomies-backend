package models

import (
	"context"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetNearGymsFromID returns the gym with the given ID if it exists or an error otherwise
func GetGymFromID(gymID string) (g interfaces.Gym, e error) {
	mongoId, err := primitive.ObjectIDFromHex(gymID)
	if err != nil {
		return interfaces.Gym{}, err
	}

	err = GymsCollection.FindOne(context.Background(), bson.M{"_id": mongoId}).Decode(&g)
	return g, err
}

// RegisterClaimedReward adds the user to the list of users that have claimed the reward for the given gym
func RegisterClaimedReward(gym interfaces.Gym, userID primitive.ObjectID) error {
	_, err := GymsCollection.UpdateOne(context.Background(), bson.M{"_id": gym.Id}, bson.M{"$push": bson.M{"rewards_claimed_by": userID}})
	return err
}

// HasUserClaimedReward returns if the user has already claimed the reward for the given gym
func HasUserClaimedReward(gym interfaces.Gym, userID primitive.ObjectID) bool {
	for _, userId := range gym.RewardsClaimedBy {
		if userId == userID {
			return true
		}
	}

	return false
}
