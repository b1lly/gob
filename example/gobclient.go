package main

import "github.com/b1lly/gob/agent"

func main() {
	agent := agent.NewGobAgent()
	agent.NewClient()

	agent.RegisterHandler(route)
}
