package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	TTL      int
}

type UserData struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

func NewRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	return client
}

func SaveUserData(client *redis.Client, userID string, data UserData, ttlSeconds int) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %v", err)
	}

	err = client.Set(ctx, "user:"+userID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save user data to Redis: %v", err)
	}

	return nil
}

func GetUserData(client *redis.Client, userID string) (*UserData, error) {
	jsonData, err := client.Get(ctx, "user:"+userID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user data from Redis: %v", err)
	}

	var data UserData
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %v", err)
	}

	return &data, nil
}

func DeleteUserData(client *redis.Client, userID string) error {
	err := client.Del(ctx, "user:"+userID).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user data from Redis: %v", err)
	}
	return nil
}
