package models

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type PostgresCredentials struct {
	Username     *string
	Password     *string
	DatabaseName *string
	SSLMode      *string
}

func MustOpenPostgres(creds PostgresCredentials) *sql.DB {
	dataSourceName := ""
	if creds.Username != nil {
		dataSourceName += " user=" + *creds.Username
	}
	if creds.Password != nil {
		dataSourceName += " password=" + *creds.Password
	}
	if creds.DatabaseName != nil {
		dataSourceName += " dbname=" + *creds.DatabaseName
	}
	if creds.SSLMode != nil {
		dataSourceName += " sslmode=" + *creds.SSLMode
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
