package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/b1lly/gob/agent"
)

func main() {
	fmt.Println("Starting up server...")

	// Start up a GobAgent and register a handler.
	// GobAgent provides a way for our applications to talk
	// to the GobServer and listen for notifications.
	go agent.StartGobAgentWithFunc("9035", "9034", handleFunc)

	// Imitate server
	for {
		time.Sleep(5 * time.Second)

		if connection := rand.Intn(10); connection > 5 {
			fmt.Println("Handling requests...")
		}
	}
}

// Sample function handler that GobAgent registers for callbacks
// from GobServer notifications
func handleFunc(files []string) {
	fmt.Println("received template update notification...")
	for _, file := range files {
		fmt.Println(file)
	}
}
