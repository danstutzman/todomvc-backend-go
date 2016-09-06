package main

import (
	"./model"
	"./web"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type Body struct {
	// ResetModel is for testing purposes
	ResetModel    bool                 `json:"resetModel"`
	DeviceUid     string               `json:"deviceUid"`
	ActionsToSync []model.ActionToSync `json:"actionsToSync"`
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

func readPostgresCredentials(path string) model.PostgresCredentials {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(fmt.Errorf("Couldn't os.Open postgres_credentials: %s", err))
	}
	defer file.Close()

	creds := model.PostgresCredentials{}
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&creds); err != nil {
		log.Fatalf("Error using decoder.Decode to parse JSON at %s: %s", path, err)
	}
	return creds
}

func main() {
	args := mustParseFlags()

	var model_ model.Model
	if args.postgresCredentialsPath != "" {
		creds := readPostgresCredentials(args.postgresCredentialsPath)
		model_ = model.NewDbModel(model.MustOpenPostgres(creds))
	} else if args.inMemoryDb {
		model_ = &model.MemoryModel{
			NextDeviceId: 1,
			Devices:      []model.Device{},
		}
	} else {
		log.Fatal("Supply either -postgres_credentials_path or -in_memory_db")
	}

	if args.socketPath != "" {
		web.MustRunSocketServer(args.socketPath, model_)
	} else {
		web.MustRunWebServer(model_)
	}
}
