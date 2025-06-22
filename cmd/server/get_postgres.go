package main

import (
	"fmt"
	"log"
	"os"
)

func getPostgresDSN() string {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_BD")

	if user == "" || password == "" || host == "" || port == "" || dbname == "" {
		log.Println("нету")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}
