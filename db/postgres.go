package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func ConnectPSQL() (*sql.DB, string) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	postgresPassword := os.Getenv("PSQL_PASSWORD")
	if postgresPassword == "" {
		log.Fatal("PSQL_PASSWORD is not set in .env file")
	}
	postgresUser := os.Getenv("PSQL_USER")
	if postgresUser == "" {
		log.Fatal("PSQL_USER is not set in .env file")
	}
	postgresDBName := os.Getenv("PSQL_DBNAME")
	if postgresDBName == "" {
		log.Fatal("PSQL_DBNAME is not set in .env file")
	}
	postgresHost := os.Getenv("PSQL_HOST")
	if postgresHost == "" {
		log.Fatal("PSQL_HOST is not set in .env file")
	}
	postgresPort := os.Getenv("PSQL_PORT")
	if postgresPort == "" {
		log.Fatal("PSQL_PORT is not set in .env file")
	}

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		postgresHost, postgresPort, postgresUser, postgresPassword, postgresDBName,
	)

	var mydb *sql.DB
	maxAttempts := 10

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		mydb, err = sql.Open("postgres", psqlInfo)
		if err == nil {
			err = mydb.Ping()
		}

		if err == nil {
			return mydb, "[CONNECTED] Connected to Postgres"
		}

		log.Printf("[RETRY %d/%d] Could not connect to Postgres: %v", attempt, maxAttempts, err)
		time.Sleep(2 * time.Second)
	}

	msg := fmt.Sprintf("[ERROR] Failed to connect to Postgres after %d attempts: %v", maxAttempts, err)
	return nil, msg
}
