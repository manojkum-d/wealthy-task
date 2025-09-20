package main

import (
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
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

	// Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer rdb.Close()

	// Redis-based rate limiter
	rrl := NewRedisTokenBucketLimiter(rdb, "api_rate_limit", 3, 10*time.Second)

	// start api server
	addr := ":8080"
	if a := os.Getenv("LISTEN_ADDR"); a != "" {
		addr = a
	}
	last:=10* time.Second
	rl:= NewRateLimit(3 , last)
	srv := NewAPIServer(addr, store, rl , rrl)
	srv.Run()
}
