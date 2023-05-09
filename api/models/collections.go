package models

import (
	"github.com/PedroChaparro/loomies-backend/configuration"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserCollection *mongo.Collection = configuration.ConnectToMongoCollection("users")
var CaughtLoomiesCollection *mongo.Collection = configuration.ConnectToMongoCollection("caught_loomies")
var ZonesCollection = configuration.ConnectToMongoCollection("zones")
var BaseLoomiesCollection = configuration.ConnectToMongoCollection("base_loomies")
var WildLoomiesCollection = configuration.ConnectToMongoCollection("wild_loomies")
var GymsCollection = configuration.ConnectToMongoCollection("gyms")
var ItemsCollection = configuration.ConnectToMongoCollection("items")
var LoomballsCollection = configuration.ConnectToMongoCollection("loom_balls")
var AuthenticationCodesCollection = configuration.ConnectToMongoCollection("authentication_codes")
var LoomieTypesCollection = configuration.ConnectToMongoCollection("loomie_types")
var LoomieRaritiesCollection = configuration.ConnectToMongoCollection("loomie_rarities")
var GymsChallengesCollection = configuration.ConnectToMongoCollection("gyms_challenges_register")
