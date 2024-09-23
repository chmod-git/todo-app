package service

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/repository"
)

type TodoListService struct {
	repo repository.TodoList
}

func NewTodoListService(repo repository.TodoList) *TodoListService {
	return &TodoListService{repo}
}

func (s *TodoListService) CreateList(userId int, list todo.TodoList) (int, error) {
	return 0, nil
}

func (s *TodoListService) GetAllLists(userId int) ([]todo.TodoList, error) {
	return []todo.TodoList{}, nil
}

func (s *TodoListService) GetListById(userId, listId int) (todo.TodoList, error) {
	return todo.TodoList{}, nil
}

func (s *TodoListService) UpdateListById(userId, listId int, list todo.UpdateListInput) error {
	return nil
}

func (s *TodoListService) DeleteListById(userId, listId int) error {
	return nil
}
