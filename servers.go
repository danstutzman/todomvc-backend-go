package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/danielstutzman/todomvc-backend-go/handlers"
	"github.com/danielstutzman/todomvc-backend-go/model"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

func mustRunWebServer(model model.Model) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, model)
	})
	log.Printf("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("Error from ListenAndServe: %s", err)
	}
}

func mustRunSocketServer(socketPath string, model model.Model) {
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

			var body handlers.Body
			if err := json.Unmarshal([]byte(bodyJson), &body); err != nil {
				l.Close()
				log.Fatalf("Error parsing JSON %s: %s", bodyJson, err)
			}

			response, err := handlers.HandleBody(body, model)
			if err != nil {
				l.Close()
				log.Fatalf("Error from HandleBody: %s", err)
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

func handleRequest(writer http.ResponseWriter, request *http.Request,
	model model.Model) {
	// Set Access-Control-Allow-Origin for all requests
	writer.Header().Set("Access-Control-Allow-Origin", "*")

	switch request.Method {
	case "GET":
		writer.Write([]byte("This API expects POST requests"))
	case "OPTIONS":
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		writer.Write([]byte("OK"))
	case "POST":
		var body handlers.Body
		decoder := json.NewDecoder(request.Body)
		if err := decoder.Decode(&body); err != nil {
			http.Error(writer, fmt.Sprintf("Error parsing JSON %s: %s", request.Body, err),
				http.StatusBadRequest)
			return
		}

		response, err := handlers.HandleBody(body, model)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Error from HandleBody: %s", err),
				http.StatusBadRequest)
			return
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Error marshaling JSON %v: %s", response, err),
				http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(responseBytes)
	default:
		http.Error(writer, fmt.Sprintf("HTTP method not allowed"),
			http.StatusMethodNotAllowed)
		return
	}
}
