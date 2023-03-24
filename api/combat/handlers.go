package combat

// ######################### Combat handlers #########################
// This functions cannot be defined in the controllers package
// because in that case, the handlers cannot access the Ws* structs
// due to the circular dependency between the combat and controllers
// packages (controllers package imports combat to use the types and
// combat imports controllers to use the handlers)

// handleGreetingMessageType is an example of how to handle a message type
// NOTE: Please, remove this function in further pull request
func handleGreetingMessageType(combat *WsCombat) {
	// Do stuff... Here you can even use models functions to interact with the database
	// Send a message to the client
	combat.SendMessage(WsMessage{
		Type:    "greeting",
		Message: "Greeting message was received",
	})
}
