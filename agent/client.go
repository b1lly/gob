package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// GobAgent communicates with GobServer (from your code)
// about changes to template files
type GobAgent struct {
	// The port GobAgent binds to
	Addr string

	// The function GobAgent should execute then it receives
	// a message about updated template files
	HandleFunc func(files []string)
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
