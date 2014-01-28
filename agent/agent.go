package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type GobAgent struct {
	Addr   string
	Routes []string
}

func NewGobAgent() *GobAgent {
	return &GobAgent{
		Addr: ":9034",
	}
}

// NewServer creates a new HTTP Server to listen for
// subscribers and notifying messages. It provides a way
// to hook third party templating engines into gob
func (ga *GobAgent) NewServer() {
	http.HandleFunc("/notify", ga.NotifySubscribers)
	http.HandleFunc("/subscribe", ga.AddRoute)

	err := http.ListenAndServe(ga.Addr, nil)
	if err != nil {
		log.Fatal("ListenAndServ: ", err)
	}
}

// AddRoute will register a particular route with the GobAgent to be
// notified when a template gets re-rendered
func (ga *GobAgent) AddRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Println(err)
		}

		err = json.Unmarshal(body, &ga.Routes)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("[gob] added subscriber to notify about template update...")
	} else {
		http.Error(w, "Post requests only.", 405)
	}
}

// NotifiySubscribers will look through the list of Routes and send them each
// a POST request with a JSON body that includes the list of source files
// that need to be rerendered
func (ga *GobAgent) NotifySubscribers(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		// Only do work if we have subscribers
		if len(ga.Routes) > 0 {
			client := &http.Client{}
			// For now, just attempt to make the POST
			// and ignore all failures
			for _, route := range ga.Routes {
				fmt.Println(route)
				clientReq, _ := http.NewRequest("POST", route, req.Body)

				_, _ = client.Do(clientReq)
			}
		} else {
			fmt.Println("[gob] please hook into the gob agent for template rendering...")
		}
	} else {
		http.Error(w, "Post requests only.", 405)
	}
}
