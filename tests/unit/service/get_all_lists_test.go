package service

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockTodoListRepository struct {
	mock.Mock
}

func (m *MockTodoListRepository) Create(userId int, list todo.TodoList) (int, error) {
	args := m.Called(userId, list)
	return args.Int(0), args.Error(1)
}

func (m *MockTodoListRepository) GetAll(userId int) ([]todo.TodoList, error) {
	args := m.Called(userId)
	return args.Get(0).([]todo.TodoList), args.Error(1)
}

func (m *MockTodoListRepository) GetById(userId, listId int) (todo.TodoList, error) {
	args := m.Called(userId, listId)
	return args.Get(0).(todo.TodoList), args.Error(1)
}

func (m *MockTodoListRepository) Update(userId, listId int, input todo.UpdateListInput) error {
	return m.Called(userId, listId, input).Error(0)
}

func (m *MockTodoListRepository) Delete(userId, listId int) error {
	return m.Called(userId, listId).Error(0)
}

func TestGetAllLists(t *testing.T) {
	mockRepo := new(MockTodoListRepository)
	service := service.NewTodoListService(mockRepo)

	userID := 1
	mockLists := []todo.TodoList{
		{Id: 1, Title: "List 1"},
		{Id: 2, Title: "List 2"},
	}
	mockRepo.On("GetAll", userID).Return(mockLists, nil)

	lists, err := service.GetAllLists(userID)
	assert.NoError(t, err)
	assert.Len(t, lists, 2)
	assert.Equal(t, "List 1", lists[0].Title)
	assert.Equal(t, "List 2", lists[1].Title)
	mockRepo.AssertExpectations(t)
}
