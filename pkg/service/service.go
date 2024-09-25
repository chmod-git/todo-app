package service

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/repository"
)

type Authorization interface {
	CreateUser(user todo.User) (int, error)
	GenerateToken(username, password string) (string, error)
	ParseToken(token string) (int, error)
}

type TodoList interface {
	CreateList(userId int, list todo.TodoList) (int, error)
	GetAllLists(userId int) ([]todo.TodoList, error)
	GetListById(userId, listId int) (todo.TodoList, error)
	UpdateListById(userId, listId int, input todo.UpdateListInput) error
	DeleteListById(userId, listId int) error
}

type TodoItem interface {
	CreateItem(userId, listId int, item todo.TodoItem) (int, error)
	GetAllItems(userId, listId int) ([]todo.TodoItem, error)
	GetItemById(userId, listId, itemId int) (todo.TodoItem, error)
	UpdateItemById(userId, listId, itemId int, input todo.UpdateItemInput) error
	DeleteItemById(userId, listId, itemId int) error
}

type Service struct {
	Authorization
	TodoList
	TodoItem
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		TodoList:      NewTodoListService(repos.TodoList),
		TodoItem:      NewTodoItemService(repos.TodoItem),
	}
}
