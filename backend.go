package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/danielstutzman/todomvc-backend-go/models"
	"log"
	"os"
)

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

func readPostgresCredentials(path string) models.PostgresCredentials {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(fmt.Errorf("Couldn't os.Open postgres_credentials: %s", err))
	}
	defer file.Close()

	creds := models.PostgresCredentials{}
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&creds); err != nil {
		log.Fatalf("Error using decoder.Decode to parse JSON at %s: %s", path, err)
	}
	return creds
}

func main() {
	args := mustParseFlags()

	var model models.Model
	if args.postgresCredentialsPath != "" {
		creds := readPostgresCredentials(args.postgresCredentialsPath)
		model = models.NewDbModel(models.MustOpenPostgres(creds))
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
