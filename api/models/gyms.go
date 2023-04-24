package models

import (
	"context"
	"fmt"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
func HasUserClaimedReward(claimed []primitive.ObjectID, userID primitive.ObjectID) bool {
	for _, userId := range claimed {
		if userId == userID {
			return true
		}
	}

	return false
}

// GetPopulatedGymFromId Returnd the details for the `/gym/:id` endpoint from the given gym id
func GetPopulatedGymFromId(GymId, UserId primitive.ObjectID) (gym interfaces.PopulatedGym, err error) {
	var auxiliarGymDoc interfaces.PopulatedGymAux
	var GymDoc interfaces.PopulatedGym

	// Populate the owner collection
	lookupIntoUsers := bson.M{
		"$lookup": bson.M{
			"from":         "users",
			"localField":   "owner",
			"foreignField": "_id",
			"as":           "owner",
		},
	}

	// Populate the loomies collection
	lookupIntoLoomies := bson.M{
		"$lookup": bson.M{
			"from":         "caught_loomies",
			"localField":   "protectors",
			"foreignField": "_id",
			"as":           "protectors",
		},
	}

	// Query the database
	cursor, err := GymsCollection.Aggregate(context.Background(), []bson.M{
		{"$match": bson.M{"_id": GymId}},
		lookupIntoUsers,
		lookupIntoLoomies,
	})

	if err != nil {
		return interfaces.PopulatedGym{}, err
	}

	// Decode the result (There is only one gym)
	if cursor.Next(context.Background()) {
		err = cursor.Decode(&auxiliarGymDoc)

		if err != nil {
			return interfaces.PopulatedGym{}, err
		}
	} else {
		return interfaces.PopulatedGym{}, fmt.Errorf("EMPTY_RESULTS")
	}

	// Parse the auxiliar gym into a populated gym
	GymDoc = *auxiliarGymDoc.ToPopulatedGym()
	GymDoc.WasRewardClaimed = HasUserClaimedReward(auxiliarGymDoc.RewardsClaimedBy, UserId)
	return GymDoc, nil
}

// GetLastGymChallengeTimestamp returns the last challenge timestamp for the given gym and player
func GetLastGymChallengeTimestamp(gymId, playerId primitive.ObjectID) (challenge interfaces.GymChallengesRegister, err error) {
	var register interfaces.GymChallengesRegister
	err = GymsChallengesCollection.FindOne(context.Background(), bson.M{"gym_id": gymId, "attacker_id": playerId}).Decode(&register)
	return register, err
}

// UpdteLastGymChallengeTimestamp updates the last challenge timestamp for the given gym and player
func UpdateLastGymChallengeTimestamp(gymId, playerId primitive.ObjectID) (err error) {
	// Check if the register exists
	currentTimestamp := time.Now().Unix()
	var register interfaces.GymChallengesRegister
	err = GymsChallengesCollection.FindOne(context.Background(), bson.M{"gym_id": gymId, "attacker_id": playerId}).Decode(&register)

	if err != nil && err == mongo.ErrNoDocuments {
		// If the register does not exist, create it
		_, err = GymsChallengesCollection.InsertOne(context.Background(), interfaces.GymChallengesRegister{
			GymId:      gymId,
			AttackerId: playerId,
			Timestamp:  currentTimestamp,
			IsActive:   true,
		})

		return err
	} else {
		// If the register exists, update it
		_, err = GymsChallengesCollection.UpdateOne(context.Background(), bson.M{"gym_id": gymId, "attacker_id": playerId}, bson.M{"$set": bson.M{"timestamp": currentTimestamp}})
		return err
	}
}

// FinishGymChallenge marks the gym challenge as finished
func FinishGymChallenge(gymId, playerId primitive.ObjectID) (err error) {
	_, err = GymsChallengesCollection.UpdateOne(context.Background(), bson.M{"gym_id": gymId, "attacker_id": playerId}, bson.M{"$set": bson.M{"is_active": false}})
 	return err
}

// UpdateGymProtectors Updates Gym Protectors (the loomies team of the owner) and new owner
func UpdateGymProtectorsAndOwner(GymId primitive.ObjectID, loomiesProtectorsIds []primitive.ObjectID, newOwner primitive.ObjectID) (err error) {
	_, err = GymsCollection.UpdateOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: GymId}},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "protectors", Value: loomiesProtectorsIds},
				{Key: "owner", Value: newOwner},
			}},
		},
	)
	return err
}
