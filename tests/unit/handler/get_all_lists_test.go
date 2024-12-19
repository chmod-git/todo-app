package handler

import (
	"encoding/json"
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

type MockTodoListService struct {
	mock.Mock
}

func (m *MockTodoListService) CreateList(userId int, list todo.TodoList) (int, error) {
	args := m.Called(userId, list)
	return args.Int(0), args.Error(1)
}

func (m *MockTodoListService) GetAllLists(userId int) ([]todo.TodoList, error) {
	args := m.Called(userId)
	return args.Get(0).([]todo.TodoList), args.Error(1)
}

func (m *MockTodoListService) GetListById(userId, listId int) (todo.TodoList, error) {
	args := m.Called(userId, listId)
	return args.Get(0).(todo.TodoList), args.Error(1)
}

func (m *MockTodoListService) UpdateListById(userId, listId int, input todo.UpdateListInput) error {
	args := m.Called(userId, listId, input)
	return args.Error(0)
}

func (m *MockTodoListService) DeleteListById(userId, listId int) error {
	args := m.Called(userId, listId)
	return args.Error(0)
}

func TestGetAllListsHandler(t *testing.T) {
	mockService := new(MockTodoListService)
	handler := handler.NewHandler(&service.Service{nil, mockService, nil})

	router := gin.Default()
	router.GET("/api/lists", handler.GetAllLists)

	userID := 1
	lists := []todo.TodoList{
		{Id: 1, Title: "List 1"},
		{Id: 2, Title: "List 2"},
	}
	mockService.On("GetAllLists", userID).Return(lists, nil)

	req, _ := http.NewRequest("GET", "/api/lists", nil)
	req.Header.Set("Authorization", "Bearer mocked_token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []todo.TodoList
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 2)
	assert.Equal(t, "List 1", response[0].Title)
	mockService.AssertExpectations(t)
}
