package service

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/repository"
)

type TodoItemService struct {
	repo repository.TodoItem
}

func NewTodoItemService(repo repository.TodoItem) *TodoItemService {
	return &TodoItemService{repo}
}

func (s *TodoItemService) CreateItem(userId, listId int, item todo.TodoItem) (int, error) {
	return s.repo.Create(userId, listId, item)
}

func (s *TodoItemService) GetAllItems(userId, listId int) ([]todo.TodoItem, error) {
	return s.repo.GetAll(userId, listId)
}

func (s *TodoItemService) GetItemById(userId, listId, itemId int) (todo.TodoItem, error) {
	return s.repo.GetById(userId, listId, itemId)
}

func (s *TodoItemService) UpdateItemById(userId, listId, itemId int, input todo.UpdateItemInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	return s.repo.Update(userId, listId, itemId, input)
}

func (s *TodoItemService) DeleteItemById(userId, listId, itemId int) error {
	return s.repo.Delete(userId, listId, itemId)
}
