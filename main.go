package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// read env

	 if err := godotenv.Load(); err != nil {
        log.Println("⚠️ No .env file found, falling back to system env vars")
    }

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://localhost:5432/postgres?sslmode=disable"
	}

	store, err := NewPostgresStore(connStr)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer store.Close()

	// start api server
	addr := ":8080"
	if a := os.Getenv("LISTEN_ADDR"); a != "" {
		addr = a
	}
	last:=5* time.Second
	rl:= NewRateLimit(last , 3)
	srv := NewAPIServer(addr, store, rl)
	srv.Run()
}
