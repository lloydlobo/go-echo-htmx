package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/lloydlobo/go-echo-htmx/internal"
)

func main() {
	runMain()
}

func runMain() {
	fmt.Println("Hello, World!")
	fmt.Printf("internal.ServerConfig: %+v\n", internal.ServerConfig)

	contacts := InitializeModels()
	fmt.Printf("contacts: %v\n", contacts)
	contacts = append(contacts, Contact{
		ID:     uuid.New(),
		Name:   "",
		Email:  "",
		Phone:  "",
		Status: StatusActive,
	})
	fmt.Printf("contacts: %v\n", contacts)

	SetupRoutes()

	// Start server
	// ...
}

func InitializeModels() []Contact {
	contacts := []Contact{}
	return contacts
}
