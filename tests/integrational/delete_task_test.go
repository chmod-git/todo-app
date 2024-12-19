package integrational

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TodoItemRepository struct {
	DB *sql.DB
}

func NewTodoItemRepository(db *sql.DB) *TodoItemRepository {
	return &TodoItemRepository{
		DB: db,
	}
}

type TodoItem struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (r *TodoItemRepository) Create(userID, listID int, item TodoItem) (int, error) {
	query := `INSERT INTO todo_items (user_id, list_id, title, description) VALUES (?, ?, ?, ?)`
	result, err := r.DB.Exec(query, userID, listID, item.Title, item.Description)
	if err != nil {
		log.Println("Error inserting todo item:", err)
		return 0, err
	}

	itemID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error retrieving last insert ID:", err)
		return 0, err
	}

	return int(itemID), nil
}

func (r *TodoItemRepository) GetById(userID, listID, itemID int) (TodoItem, error) {
	var item TodoItem
	query := `SELECT id, title, description FROM todo_items WHERE user_id = ? AND list_id = ? AND id = ?`
	row := r.DB.QueryRow(query, userID, listID, itemID)
	if err := row.Scan(&item.ID, &item.Title, &item.Description); err != nil {
		if err == sql.ErrNoRows {
			return TodoItem{}, errors.New("item not found")
		}
		log.Println("Error getting todo item:", err)
		return TodoItem{}, err
	}
	return item, nil
}

func (r *TodoItemRepository) Delete(userID, listID, itemID int) error {
	query := `DELETE FROM todo_items WHERE user_id = ? AND list_id = ? AND id = ?`
	_, err := r.DB.Exec(query, userID, listID, itemID)
	if err != nil {
		log.Println("Error deleting todo item:", err)
		return err
	}
	return nil
}

func TestDeleteTaskIntegration(t *testing.T) {
	db := InitTestDB()
	redisClient := InitTestRedis()
	repo := NewRepository(db, redisClient)
	service := NewService(repo)
	handler := NewHandler(service.Authorization, service.TodoList, service.TodoItem)

	router := gin.Default()
	InitRoutes(router, handler)

	token := getAuthTokenForIntegrationTest(router)

	userID := 1
	listID, _ := repo.TodoList.Create(userID, todo.TodoList{Title: "Test List"})
	itemID, _ := repo.TodoItem.Create(userID, listID, TodoItem{Title: "Test Item", Description: "Test Description"})

	item, err := repo.TodoItem.GetById(userID, listID, itemID)
	assert.NoError(t, err)
	assert.Equal(t, itemID, item.ID)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/lists/%d/items/%d", listID, itemID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "item deleted", response["message"])

	_, err = repo.TodoItem.GetById(userID, listID, itemID)
	assert.Error(t, err)
}
