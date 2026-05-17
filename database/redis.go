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
	dsn := os.Getenv("REDIS_URL") // Contoh: redis://:password@localhost:6379/0
	if dsn == "" {
		dsn = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(dsn)
	if err != nil {
		log.Fatalf("Gagal memparsing REDIS_URL: %v", err)
	}

	RDB = redis.NewClient(opt)

	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Gagal terhubung ke Redis: %v", err)
	}
	log.Println("Koneksi Redis berhasil!")
}
