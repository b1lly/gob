package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/b1lly/gob/agent"
)

func main() {
	fmt.Println("Starting up server...")
	go agent.StartGobAgentWithFunc("9035", "9034", handleFunc)

	// Imitate server
	for {
		time.Sleep(1 * time.Second)

		if connection := rand.Intn(10); connection > 5 {
			// fmt.Println("Sample application...")
		}
	}
}

func handleFunc(files []string) {
	fmt.Println("received template update notification...")
	for _, file := range files {
		fmt.Println(file)
	}
}
