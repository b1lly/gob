package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// GobServer represents the single server gob
// runs to notify subscribers of changes to
// template files (currently only soy)
type GobServer struct {
	// The port GobServer binds to
	Addr string

	// The hosts GobServer should update about template changes
	SubscriberRoutes []string
}

// Creates a new GobServer which will bind to the specified port
func NewGobServer(port string) *GobServer {
	return &GobServer{
		Addr: fmt.Sprintf(":%s", port),
	}
}

// NewGobServer creates a new HTTP Server to listen for
// subscribers and notifying messages. It provides a way
// to hook third party templating engines into gob
func (ga *GobServer) Start() {
	http.HandleFunc("/subscribe", ga.AddRoute)

	fmt.Printf("[gob] starting up server on port %s\n", ga.Addr)
	err := http.ListenAndServe(ga.Addr, nil)
	if err != nil {
		log.Fatal("ListenAndServ: ", err)
	}
}

// AddRoute will register a particular route with the GobAgent to be
// notified when a template gets re-rendered
func (ga *GobServer) AddRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		data := make(map[string]string)
		err = json.Unmarshal(body, &data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ga.SubscriberRoutes = append(ga.SubscriberRoutes, data["route"])
		fmt.Println("[gob] added subscriber to notify about template update...")
	} else {
		http.Error(w, "Post requests only.", 405)
	}
}

// NotifiySubscribers will look through the list of Routes and send them each
// a POST request with a JSON body that includes the list of source files
// that need to be rerendered
func (ga *GobServer) NotifySubscribers(files []string) {
	// Only do work if we have subscribers
	if len(ga.SubscriberRoutes) > 0 {
		client := &http.Client{}
		// For now, just attempt to make the POST
		// and ignore all failures
		fileMap := map[string][]string{
			"files": files,
		}
		data, err := json.Marshal(&fileMap)
		if err != nil {
			panic(err)
		}
		for _, route := range ga.SubscriberRoutes {
			//TODO(lt)
			clientReq, err := http.NewRequest("POST", "http://"+route, bytes.NewReader(data))
			if err != nil {
				panic(err)
			}

			// TODO(ttacon): check for errors
			_, _ = client.Do(clientReq)
		}
	} else {
		fmt.Println("[gob] please hook into the gob agent for template rendering...")
	}
}
