package handler

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/handler"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockTodoItemService struct {
	mock.Mock
}

func (m *MockTodoItemService) CreateItem(userId, listId int, item todo.TodoItem) (int, error) {
	args := m.Called(userId, listId, item)
	return args.Int(0), args.Error(1)
}

func (m *MockTodoItemService) GetAllItems(userId, listId int) ([]todo.TodoItem, error) {
	args := m.Called(userId, listId)
	return args.Get(0).([]todo.TodoItem), args.Error(1)
}

func (m *MockTodoItemService) GetItemById(userId, listId, itemId int) (todo.TodoItem, error) {
	args := m.Called(userId, listId, itemId)
	return args.Get(0).(todo.TodoItem), args.Error(1)
}

func (m *MockTodoItemService) UpdateItemById(userId, listId, itemId int, input todo.UpdateItemInput) error {
	args := m.Called(userId, listId, itemId, input)
	return args.Error(0)
}

func (m *MockTodoItemService) DeleteItemById(userId, listId, itemId int) error {
	args := m.Called(userId, listId, itemId)
	return args.Error(0)
}

func TestDeleteItemHandler(t *testing.T) {
	mockService := new(MockTodoItemService)
	handler := handler.NewHandler(&service.Service{nil, nil, mockService})

	router := gin.Default()
	router.DELETE("/api/lists/:list_id/items/:item_id", handler.DeleteItemById)

	userID := 1
	listID := 1
	itemID := 42
	mockService.On("DeleteItemById", userID, listID, itemID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/lists/1/items/42", nil)
	req.Header.Set("Authorization", "Bearer mocked_token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
