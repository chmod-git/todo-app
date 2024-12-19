package service

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockAuthorizationRepository struct {
	mock.Mock
}

func (m *MockAuthorizationRepository) GetUser(username, password string) (todo.User, error) {
	args := m.Called(username, password)
	return args.Get(0).(todo.User), args.Error(1)
}

func (m *MockAuthorizationRepository) UpdateUser(userId int, user todo.User) error {
	return m.Called(userId, user).Error(0)
}

func (m *MockAuthorizationRepository) CreateUser(user todo.User) (int, error) {
	args := m.Called(user)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthorizationRepository) DeleteUser(userId int) error {
	return m.Called(userId).Error(0)
}

func TestGenerateToken(t *testing.T) {
	mockRepo := new(MockAuthorizationRepository)
	service := service.NewAuthService(mockRepo)

	username := "testuser"
	password := "password"
	mockRepo.On("GetUser", username, password).Return(todo.User{Id: 1, Username: username}, nil)

	token, err := service.GenerateToken(username, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}
