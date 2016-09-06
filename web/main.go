package web

import (
	"bufio"
	"encoding/json"
	"github.com/danielstutzman/todomvc-backend-go/model"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

type Body struct {
	// ResetModel is for testing purposes
	ResetModel    bool                 `json:"resetModel"`
	DeviceUid     string               `json:"deviceUid"`
	ActionsToSync []model.ActionToSync `json:"actionsToSync"`
}

func MustRunWebServer(model model.Model) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, model)
	})
	log.Printf("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("Error from ListenAndServe: %s", err)
	}
}

func MustRunSocketServer(socketPath string, model model.Model) {
	log.Printf("Listening on %s...", socketPath)
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()

	// Shut down server (delete socket file) if SIGINT received
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		l.Close()
		os.Exit(2)
	}(sigc)

	for {
		fd, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			bodyJson := scanner.Text()

			var body Body
			if err := json.Unmarshal([]byte(bodyJson), &body); err != nil {
				l.Close()
				log.Fatalf("Error parsing JSON %s: %s", bodyJson, err)
			}

			response, err := handleBody(body, model)
			if err != nil {
				l.Close()
				log.Fatalf("Error from handleBody: %s", err)
			}

			responseJson, err := json.Marshal(response)
			if err != nil {
				l.Close()
				log.Fatalf("Error marshaling JSON %v: %s", response, err)
			}
			log.Printf("Response: %s", responseJson)

			_, err = fd.Write(responseJson)
			if err != nil {
				l.Close()
				log.Fatal("Error from Write: ", err)
			}

			_, err = fd.Write([]byte("\n"))
			if err != nil {
				l.Close()
				log.Fatal("Error from Write: ", err)
			}
		} // scan next line
	} // endless loop of accepting more connections

} // end mustRunSocketServer
