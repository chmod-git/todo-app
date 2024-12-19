package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisRepository struct {
	listClient *redis.Client
	itemClient *redis.Client
}

func NewRedisRepository(listDB, itemDB int, host, port, password string) *RedisRepository {
	listClient := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       listDB,
	})

	itemClient := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       itemDB,
	})

	ctx := context.Background()
	if _, err := listClient.Ping(ctx).Result(); err != nil {
		panic("Failed to connect to Redis (lists): " + err.Error())
	}
	if _, err := itemClient.Ping(ctx).Result(); err != nil {
		panic("Failed to connect to Redis (items): " + err.Error())
	}

	return &RedisRepository{
		listClient: listClient,
		itemClient: itemClient,
	}
}

var ctx = context.Background()

func (r *RedisRepository) SaveListsData(chatID string, data []todo.TodoList, ttlSeconds int) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal list data: %v", err)
	}

	err = r.listClient.Set(ctx, "chat:"+chatID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save list data to Redis: %v", err)
	}

	return nil
}

func (r *RedisRepository) GetListsData(chatID string) ([]todo.TodoList, error) {
	jsonData, err := r.listClient.Get(ctx, "chat:"+chatID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("no data found for chat ID: %s", chatID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get list data from Redis: %v", err)
	}

	var data []todo.TodoList
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal list data: %v", err)
	}

	return data, nil
}

func (r *RedisRepository) UpdateListData(chatID string, updatedList todo.TodoList, ttlSeconds int) error {
	data, err := r.GetListsData(chatID)
	if err != nil {
		return fmt.Errorf("failed to get list data for update: %v", err)
	}

	updated := false

	for i, list := range data {
		if list.Id == updatedList.Id {
			data[i] = updatedList
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("list with ID %v not found", updatedList.Id)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal updated list data: %v", err)
	}

	err = r.listClient.Set(ctx, "chat:"+chatID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save updated list data to Redis: %v", err)
	}

	return nil
}

func (r *RedisRepository) AddListData(chatID string, newList todo.TodoList, ttlSeconds int) error {
	data, err := r.GetListsData(chatID)
	if err != nil {
		return fmt.Errorf("failed to get list data for addition: %v", err)
	}

	for _, list := range data {
		if list.Id == newList.Id {
			return fmt.Errorf("list with ID %v already exists", newList.Id)
		}
	}

	data = append(data, newList)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal new list data: %v", err)
	}

	err = r.listClient.Set(ctx, "chat:"+chatID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save new list data to Redis: %v", err)
	}

	return nil
}

func (r *RedisRepository) DeleteListData(chatID string, targetList todo.TodoList, ttlSeconds int) error {
	data, err := r.GetListsData(chatID)
	if err != nil {
		return fmt.Errorf("failed to get list data for deletion: %v", err)
	}

	var updatedData []todo.TodoList
	for _, list := range data {
		if list.Id != targetList.Id {
			updatedData = append(updatedData, list)
		}
	}

	jsonData, err := json.Marshal(updatedData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated list data: %v", err)
	}

	err = r.listClient.Set(ctx, "chat:"+chatID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save updated list data to Redis: %v", err)
	}

	return nil
}

func (r *RedisRepository) SaveItemsData(chatID string, data []todo.TodoItem, ttlSeconds int) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal item data: %v", err)
	}

	err = r.itemClient.Set(ctx, "chat:"+chatID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save item data to Redis: %v", err)
	}

	return nil
}

func (r *RedisRepository) GetItemsData(chatID string) ([]todo.TodoItem, error) {
	jsonData, err := r.itemClient.Get(ctx, "chat:"+chatID).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("no data found for chat ID: %s", chatID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get item data from Redis: %v", err)
	}

	var data []todo.TodoItem
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal list data: %v", err)
	}

	return data, nil
}

func (r *RedisRepository) UpdateItemData(chatID string, updatedItem todo.TodoItem, ttlSeconds int) error {
	data, err := r.GetItemsData(chatID)
	if err != nil {
		return fmt.Errorf("failed to get item data for update: %v", err)
	}

	updated := false

	for i, item := range data {
		if item.Id == updatedItem.Id {
			data[i] = updatedItem
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("item with ID %v not found", updatedItem.Id)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal updated item data: %v", err)
	}

	err = r.itemClient.Set(ctx, "chat:"+chatID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save updated item data to Redis: %v", err)
	}

	return nil
}

func (r *RedisRepository) AddItemData(chatID string, newItem todo.TodoItem, ttlSeconds int) error {
	data, err := r.GetItemsData(chatID)
	if err != nil {
		return fmt.Errorf("failed to get item data for addition: %v", err)
	}

	for _, item := range data {
		if item.Id == newItem.Id {
			return fmt.Errorf("item with ID %v already exists", newItem.Id)
		}
	}

	data = append(data, newItem)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal new item data: %v", err)
	}

	err = r.itemClient.Set(ctx, "chat:"+chatID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save new item data to Redis: %v", err)
	}

	return nil
}

func (r *RedisRepository) DeleteItemData(chatID string, targetItem todo.TodoItem, ttlSeconds int) error {
	data, err := r.GetItemsData(chatID)
	if err != nil {
		return fmt.Errorf("failed to get list data for deletion: %v", err)
	}

	var updatedData []todo.TodoItem
	for _, item := range data {
		if item.Id != targetItem.Id {
			updatedData = append(updatedData, item)
		}
	}

	jsonData, err := json.Marshal(updatedData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated item data: %v", err)
	}

	err = r.itemClient.Set(ctx, "chat:"+chatID, jsonData, time.Duration(ttlSeconds)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to save updated item data to Redis: %v", err)
	}

	return nil
}
