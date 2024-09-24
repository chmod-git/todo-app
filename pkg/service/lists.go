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
	return s.repo.Create(userId, list)
}

func (s *TodoListService) GetAllLists(userId int) ([]todo.TodoList, error) {
	return s.repo.GetAll(userId)
}

func (s *TodoListService) GetListById(userId, listId int) (todo.TodoList, error) {
	return s.repo.GetById(userId, listId)
}

func (s *TodoListService) UpdateListById(userId, listId int, input todo.UpdateListInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	return s.repo.Update(userId, listId, input)
}

func (s *TodoListService) DeleteListById(userId, listId int) error {
	return s.repo.Delete(userId, listId)
}
