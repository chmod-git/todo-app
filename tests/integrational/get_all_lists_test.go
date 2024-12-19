package integrational

import (
	"encoding/json"
	"github.com/chmod-git/todo-app"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllListsIntegration(t *testing.T) {
	db := InitTestDB()
	redisClient := InitTestRedis()
	repo := NewRepository(db, redisClient)
	service := NewService(repo)
	handler := NewHandler(service.Authorization, service.TodoList, service.TodoItem)

	router := gin.Default()
	InitRoutes(router, handler)

	token := getAuthTokenForIntegrationTest(router)

	userID := 1
	repo.TodoList.Create(userID, todo.TodoList{Title: "Test List 1"})
	repo.TodoList.Create(userID, todo.TodoList{Title: "Test List 2"})

	req, _ := http.NewRequest("GET", "/api/lists", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []todo.TodoList
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 2)
	assert.Equal(t, "Test List 1", response[0].Title)
	assert.Equal(t, "Test List 2", response[1].Title)
}
