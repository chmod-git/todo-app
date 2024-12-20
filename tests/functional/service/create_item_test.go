package service

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockTodoItemRepository struct {
	mock.Mock
}

func (m *MockTodoItemRepository) Create(userId, listId int, item todo.TodoItem) (int, error) {
	args := m.Called(userId, listId, item)
	return args.Int(0), args.Error(1)
}

func (m *MockTodoItemRepository) GetAll(userId, listId int) ([]todo.TodoItem, error) {
	args := m.Called(userId, listId)
	return args.Get(0).([]todo.TodoItem), args.Error(1)
}

func (m *MockTodoItemRepository) GetById(userId, listId, itemId int) (todo.TodoItem, error) {
	args := m.Called(userId, listId, itemId)
	return args.Get(0).(todo.TodoItem), args.Error(1)
}

func (m *MockTodoItemRepository) Update(userId, listId, itemId int, input todo.UpdateItemInput) error {
	return m.Called(userId, listId, itemId, input).Error(0)
}

func (m *MockTodoItemRepository) Delete(userId, listId, itemId int) error {
	return m.Called(userId, listId, itemId).Error(0)
}

func TestCreateItem(t *testing.T) {
	mockRepo := new(MockTodoItemRepository)
	service := service.NewTodoItemService(mockRepo)

	userID := 1
	listID := 1
	item := todo.TodoItem{Title: "New Task", Description: "Task Description"}
	mockRepo.On("Create", userID, listID, item).Return(42, nil)

	itemID, err := service.CreateItem(userID, listID, item)
	assert.NoError(t, err)
	assert.Equal(t, 42, itemID)
	mockRepo.AssertExpectations(t)
}
