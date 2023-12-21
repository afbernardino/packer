package main

import (
	"database/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"log"
	"os"
	"packer/internal/rest"
	"packer/internal/rest/order/repository"
	"strconv"
	"time"
)

const (
	envDatabaseURL             = "DATABASE_URL"
	envDatabaseConnMaxLifeTime = "DATABASE_CONNECTION_MAX_LIFE_TIME"
	envReadTimeoutSeconds      = "READ_TIMEOUT_SECONDS"
	envWriteTimeoutSeconds     = "WRITE_TIMEOUT_SECONDS"
	envIdleTimeoutSeconds      = "IDLE_TIMEOUT_SECONDS"
	envPort                    = "PORT"
)

func main() {
	db, err := sql.Open("pgx", newDbConnString())
	if err != nil {
		log.Fatalf("unable to connect to database: %v\n", err)
	}

	defer func(db *sql.DB) {
		closeErr := db.Close()
		if closeErr != nil {
			log.Fatalf("unable to close the database connection: %v\n", err)
		}
	}(db)

	setDbConnMaxLifeTime(db)
	pingDb(db)

	repo := repository.NewDatabase(db)
	svc := rest.NewApiService(newRestApiConfig(), &repo)

	port := getEnvIntOrDefault(envPort, 8080)
	svc.Serve(port)
}

func newDbConnString() string {
	dbUrl := os.Getenv(envDatabaseURL)
	if dbUrl == "" {
		log.Fatal("database url has not been set")
	}

	connConfig, err := pgx.ParseConfig(dbUrl)
	if err != nil {
		log.Fatalf("unable to parse database config: %v\n", err)
	}

	return stdlib.RegisterConnConfig(connConfig)
}

func setDbConnMaxLifeTime(db *sql.DB) {
	dbConnMaxLifeTime := getEnvIntOrDefault(envDatabaseConnMaxLifeTime, 90)
	db.SetConnMaxLifetime(time.Duration(dbConnMaxLifeTime) * time.Second)
}

func pingDb(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatalf("unable to ping the database: %v\n", err)
	}
}

func newRestApiConfig() rest.Config {
	readTimeout := getEnvIntOrDefault(envReadTimeoutSeconds, 30)
	writeTimeout := getEnvIntOrDefault(envWriteTimeoutSeconds, 90)
	idleTimeout := getEnvIntOrDefault(envIdleTimeoutSeconds, 120)
	return rest.Config{
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
	}
}

func getEnvIntOrDefault(key string, def int) int {
	value := os.Getenv(key)
	if value == "" {
		return def
	}

	atoi, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("the following environment variable should be an int: %s\n", key)
		log.Printf("using the default value: %d\n", def)
		return def
	}
	return atoi
}
