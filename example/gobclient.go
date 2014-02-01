package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	// Route to notify me on (this can be a list of items)
	// You can have multiple subscribers
	routes := []string{"http://localhost:3000/templateupdate"}

	// The POST body requires a JSON object
	enc, err := json.Marshal(routes)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Subscribe to be notified by the Gob Agent when any views/templates
	// are modified/added/deleted
	client := &http.Client{}
	clientReq, _ := http.NewRequest("POST", "http://localhost:9034", bytes.NewReader(enc))
	res, err := client.Do(clientReq)

	if err != nil {
		fmt.Println(err)
	}

	// The response sends back a list of files in the body
	// So we can unmarshal it from the response into a slice of strings
	files := []string{}
	// err = json.Unmarshal(res.Body, &files)
}
