package main

import (
	"./models"
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

type Body struct {
	// ResetModel is for testing purposes
	ResetModel    bool           `json:"resetModel"`
	DeviceUid     string         `json:"deviceUid"`
	ActionsToSync []ActionToSync `json:"actionsToSync"`
}

type ActionToSync struct {
	Id int `json:"id"`
}

type CommandLineArgs struct {
	postgresCredentialsPath string
	socketPath              string
	inMemoryDb              bool
}

func mustParseFlags() CommandLineArgs {
	var args CommandLineArgs
	flag.StringVar(&args.postgresCredentialsPath, "postgres_credentials_path", "",
		"JSON file with username and password")
	flag.StringVar(&args.socketPath, "socket_path", "",
		"Path for UNIX socket server for testing")
	flag.BoolVar(&args.inMemoryDb, "in_memory_db", false,
		"Store data in memory instead of PostgreSQL for faster testing")
	flag.Parse()
	return args
}

func mustOpenPostgres(postgresCredentialsPath string) *sql.DB {
	postgresCredentialsFile, err := os.Open(postgresCredentialsPath)
	if err != nil {
		log.Fatal(fmt.Errorf("Couldn't os.Open postgres_credentials: %s", err))
	}
	defer postgresCredentialsFile.Close()

	type PostgresCredentials struct {
		Username     *string
		Password     *string
		DatabaseName *string
		SSLMode      *string
	}
	postgresCredentials := PostgresCredentials{}
	decoder := json.NewDecoder(postgresCredentialsFile)
	if err = decoder.Decode(&postgresCredentials); err != nil {
		log.Fatalf("Error using decoder.Decode to parse JSON at %s: %s",
			postgresCredentialsPath, err)
	}

	dataSourceName := ""
	if postgresCredentials.Username != nil {
		dataSourceName += " user=" + *postgresCredentials.Username
	}
	if postgresCredentials.Password != nil {
		dataSourceName += " password=" + *postgresCredentials.Password
	}
	if postgresCredentials.DatabaseName != nil {
		dataSourceName += " dbname=" + *postgresCredentials.DatabaseName
	}
	if postgresCredentials.SSLMode != nil {
		dataSourceName += " sslmode=" + *postgresCredentials.SSLMode
	}

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(fmt.Errorf("Error from sql.Open: %s", err))
	}

	// Test out the database connection immediately to check the credentials
	ignored := 0
	err = db.QueryRow("SELECT 1").Scan(&ignored)
	if err != nil {
		log.Fatal(fmt.Errorf("Error from db.QueryRow: %s", err))
	}

	return db
}

func mustRunWebServer(model models.Model) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, model)
	})
	log.Printf("Listening on :3000...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalf("Error from ListenAndServe: %s", err)
	}
}

func mustRunSocketServer(socketPath string, model models.Model) {
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
				log.Fatalf("Error parsing JSON %s: %s", bodyJson, err)
			}

			response, err := handleBody(body, model)
			if err != nil {
				log.Fatalf("Error from handleBody: %s", err)
			}

			responseJson, err := json.Marshal(response)
			if err != nil {
				log.Fatalf("Error marshaling JSON %s: %s", response, err)
			}

			_, err = fd.Write(responseJson)
			if err != nil {
				log.Fatal("Write: ", err)
			}
		} // scan next line
	} // endless loop of accepting more connections

} // end mustRunSocketServer

func main() {
	args := mustParseFlags()

	var model models.Model
	if args.postgresCredentialsPath != "" {
		model = models.NewDbModel(mustOpenPostgres(args.postgresCredentialsPath))
	} else if args.inMemoryDb {
		model = &models.MemoryModel{
			NextDeviceId: 1,
			Devices:      []models.Device{},
		}
	} else {
		log.Fatal("Supply either -postgres_credentials_path or -in_memory_db")
	}

	if args.socketPath != "" {
		mustRunSocketServer(args.socketPath, model)
	} else {
		mustRunWebServer(model)
	}
}
