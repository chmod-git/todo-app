package handler

import (
	"bytes"
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

type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) CreateUser(user todo.User) (int, error) {
	args := m.Called(user)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthorizationService) GenerateToken(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthorizationService) GetUser(username, password string) (todo.User, error) {
	args := m.Called(username, password)
	return args.Get(0).(todo.User), args.Error(1)
}

func (m *MockAuthorizationService) UpdateUser(userId int, user todo.User) error {
	args := m.Called(userId, user)
	return args.Error(0)
}

func (m *MockAuthorizationService) DeleteUser(userId int) error {
	args := m.Called(userId)
	return args.Error(0)
}

func (m *MockAuthorizationService) ParseToken(token string) (int, error) {
	args := m.Called(token)
	return args.Int(0), args.Error(1)
}

func TestSignUpHandler(t *testing.T) {
	mockService := new(MockAuthorizationService)
	handler := handler.NewHandler(&service.Service{mockService, nil, nil})

	router := gin.Default()
	router.POST("/auth/sign-up", handler.SignUp)

	user := todo.User{Username: "testuser", Password: "password"}
	mockService.On("CreateUser", user).Return(1, nil)

	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
