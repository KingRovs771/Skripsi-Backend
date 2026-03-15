package database

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	dsn := os.Getenv("REDIS_URL") // Contoh: localhost:6379
	if dsn == "" {
		dsn = "localhost:6379"
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:     dsn,
		Password: "",
		DB:       0,
	})

	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Gagal terhubung ke Redis: %v", err)
	}
	log.Println("Koneksi Redis berhasil!")
}
