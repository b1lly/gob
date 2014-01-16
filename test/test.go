package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("Starting up server on port 3000...")

	// Imitate server
	for {
		time.Sleep(3 * time.Second)

		if connection := rand.Intn(10); connection > 5 {
			fmt.Println("Handling a network good...")
			fmt.Println("Sending data.html to wire...")
		}
	}
}
