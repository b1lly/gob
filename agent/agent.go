package agent

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

// GobAgent communicates with GobServer (from your code)
// about changes to template files
type GobAgent struct {
	// The port GobAgent binds to
	Addr string

	// The function GobAgent should execute then it receives
	// a message about updated template files
	HandleFunc func(files []string)
}

// Creates a new GobServer which will bind to the specified port
func NewGobServer(port string) *GobServer {
	return &GobServer{
		Addr: fmt.Sprintf(":%s", port),
	}
}

// Creates a new GobAgent which will bind to the specified port
func NewGobAgent(port string) *GobAgent {
	return &GobAgent{
		Addr: fmt.Sprintf(":%s", port),
	}
}

// StartGobAgentWithFunc handles all the boilerplate normally required
// to start up a GobAgent
func StartGobAgentWithFunc(agentPort, serverPort string, f func([]string)) {
	ga := NewGobAgent(agentPort)
	ga.SetHandleFunc(f)
	ga.Start(serverPort)
}

// SetHandleFunc registers a function to call with the GobAgent when
// an update message is recieved. The only parameter it should take is
// the list of changed files
func (ga *GobAgent) SetHandleFunc(f func([]string)) {
	ga.HandleFunc = f
}

// Start sets up the web server for a GobAgent, subscribes to
// the GobServer (if it's up) and begins to listen for updates
func (ga *GobAgent) Start(serverPort string) {
	http.HandleFunc("/update", ga.HandleUpdate)

	if err := ga.Subscribe(serverPort); err != nil {
		fmt.Println(err)
		fmt.Println("[gob] Failed to connect to gob server, gob agent turning off...")
		return
	}

	err := http.ListenAndServe(ga.Addr, nil)
	if err != nil {
		log.Fatal("ListenAndServ: ", err)
	}
}

// HandleUpdate receives the files that have changed and
// calls the registered function on them
func (ga *GobAgent) HandleUpdate(w http.ResponseWriter, req *http.Request) {
	var files map[string][]string
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(data, &files)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO(ttacon): check for key existence
	ga.HandleFunc(files["files"])
	w.WriteHeader(http.StatusOK)
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

// Subscribe registers a GobAgent with a GobServer listening
// on the given port
func (ga *GobAgent) Subscribe(serverPort string) error {
	client := http.Client{}
	body := map[string]string{
		"route": ga.Addr + "/update",
	}
	data, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://localhost:%s/subscribe", serverPort),
		bytes.NewReader(data))
	if err != nil {
		return err
	}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
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
