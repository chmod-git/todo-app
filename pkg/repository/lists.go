package repository

import (
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/jmoiron/sqlx"
)

type ListSQL struct {
	db *sqlx.DB
}

func NewListSQL(db *sqlx.DB) *ListSQL {
	return &ListSQL{db: db}
}

func (s *ListSQL) Create(userId int, list todo.TodoList) (int, error) {
	var listId int
	query := fmt.Sprintf("INSERT INTO %s (title, description, user_id) VALUES ($1, $2, $3) RETURNING id", todoListTable)

	err := s.db.Get(&listId, query, list.Title, list.Description, userId)
	if err != nil {
		return 0, err
	}

	return listId, nil
}

func (s *ListSQL) GetAll(userId int) ([]todo.TodoList, error) {
	var lists []todo.TodoList
	query := fmt.Sprintf("SELECT tl.id, tl.title, tl.description FROM %s tl WHERE tl.user_id = $1", todoListTable)

	err := s.db.Select(&lists, query, userId)
	if err != nil {
		return nil, err
	}

	return lists, nil
}

func (s *ListSQL) GetById(userId, listId int) (todo.TodoList, error) {
	var list todo.TodoList
	query := fmt.Sprintf("SELECT tl.id, tl.title, tl.description FROM %s tl WHERE tl.user_id = $1 AND tl.id = $2", todoListTable)

	err := s.db.Get(&list, query, userId, listId)
	if err != nil {
		return todo.TodoList{}, err
	}

	return list, nil
}

func (s *ListSQL) Update(userId, listId int, input todo.UpdateListInput) error {
	query := fmt.Sprintf("UPDATE %s tl SET title = COALESCE($1, title), description = COALESCE($2, description) WHERE tl.user_id = $3 AND tl.id = $4", todoListTable)

	_, err := s.db.Exec(query, input.Title, input.Description, userId, listId)
	return err
}

func (s *ListSQL) Delete(userId, listId int) error {
	query := fmt.Sprintf("DELETE FROM %s tl WHERE tl.user_id = $1 AND tl.id = $2", todoListTable)

	_, err := s.db.Exec(query, userId, listId)
	return err
}
