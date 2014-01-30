package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type GobServer struct {
	Addr             string
	SubscriberRoutes []string
}

type GobAgent struct {
	Addr       string
	HandleFunc func(files []string)
}

func NewGobServer(port string) *GobServer {
	return &GobServer{
		Addr: fmt.Sprintf(":%s", port),
	}
}

func NewGobAgent(port string) *GobAgent {
	return &GobAgent{
		Addr: fmt.Sprintf(":%s", port),
	}
}

func StartGobAgentWithFunc(agentPort, serverPort string, f func([]string)) {
	ga := NewGobAgent(agentPort)
	ga.SetHandleFunc(f)
	ga.Start(serverPort)
}

func (ga *GobAgent) SetHandleFunc(f func([]string)) {
	ga.HandleFunc = f
}

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

func (ga *GobAgent) Subscribe(serverPort string) error {
	client := http.Client{}
	body := map[string]string{
		"route": ga.Addr + "/update",
	}
	data, err := json.Marshal(&body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/subscribe", serverPort), bytes.NewReader(data))
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
