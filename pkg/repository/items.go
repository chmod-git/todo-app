package repository

import (
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/jmoiron/sqlx"
)

type ItemSQL struct {
	db *sqlx.DB
}

func NewItemSQL(db *sqlx.DB) *ItemSQL {
	return &ItemSQL{db: db}
}

func (s *ItemSQL) Create(userId, listId int, item todo.TodoItem) (int, error) {
	exists, err := s.checkListOwnership(userId, listId)

	if err != nil {
		return 0, err
	} else if !exists {
		return 0, fmt.Errorf("list does not exist")
	}

	var itemId int
	query := fmt.Sprintf("INSERT INTO %s (title, description, done, list_id) VALUES ($1, $2, $3, $4) RETURNING id", todoItemTable)

	err = s.db.Get(&itemId, query, item.Title, item.Description, item.Done, listId)
	return itemId, err
}

func (s *ItemSQL) GetAll(userId, listId int) ([]todo.TodoItem, error) {
	exists, err := s.checkListOwnership(userId, listId)

	if err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("list does not exist")
	}

	var items []todo.TodoItem
	query := fmt.Sprintf("SELECT * FROM %s WHERE list_id = $1", todoItemTable)

	err = s.db.Select(&items, query, listId)
	return items, err
}

func (s *ItemSQL) GetById(userId, listId, itemId int) (todo.TodoItem, error) {
	exists, err := s.checkListOwnership(userId, listId)

	if err != nil {
		return todo.TodoItem{}, err
	} else if !exists {
		return todo.TodoItem{}, fmt.Errorf("list does not exist")
	}

	var item todo.TodoItem
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1 AND list_id = $2", todoItemTable)

	err = s.db.Get(&item, query, itemId, listId)
	return item, err
}

func (s *ItemSQL) Update(userId, listId, itemId int, input todo.UpdateItemInput) error {
	exists, err := s.checkListOwnership(userId, listId)

	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("list does not exist")
	}

	if err = s.existsItem(itemId, listId); err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %s SET title = COALESCE($1, title), description = COALESCE($2, description), done = COALESCE($3, done) WHERE id = $4 AND list_id = $5", todoItemTable)

	_, err = s.db.Exec(query, input.Title, input.Description, input.Done, itemId, listId)
	return err
}

func (s *ItemSQL) Delete(userId, listId, itemId int) error {
	exists, err := s.checkListOwnership(userId, listId)

	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("list does not exist")
	}

	if err = s.existsItem(itemId, listId); err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND list_id = $2", todoItemTable)

	_, err = s.db.Exec(query, itemId, listId)
	return err
}

func (s *ItemSQL) checkListOwnership(userId, listId int) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE id = $1 AND user_id = $2)", todoListTable)

	var exists bool
	err := s.db.Get(&exists, query, listId, userId)

	return exists, err
}

func (s *ItemSQL) existsItem(itemId, listId int) error {
	var existsItem bool
	err := s.db.Get(&existsItem, fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE id = $1 AND list_id = $2)", todoItemTable), itemId, listId)

	if err != nil {
		return err
	} else if !existsItem {
		return fmt.Errorf("item does not exist")
	}

	return nil
}
