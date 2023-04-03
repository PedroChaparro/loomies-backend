package models

import "github.com/PedroChaparro/loomies-backend/configuration"

var AuthenticationCodesCollection = configuration.ConnectToMongoCollection("authentication_codes")
