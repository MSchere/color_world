package main

import (
	"os"
	"strconv"

	"github.com/gofiber/storage/redis/v3"
	"github.com/joho/godotenv"
)

func RedisConnection() *redis.Storage {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	store := redis.New(redis.Config{
		Host: redisHost,
		Port: redisPort,
	})
	return store
}
