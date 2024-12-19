package repository

import (
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/jmoiron/sqlx"
)

type AuthSQL struct {
	db *sqlx.DB
}

func NewAuthSQL(db *sqlx.DB) *AuthSQL {
	return &AuthSQL{db: db}
}

func (s *AuthSQL) CreateUser(user todo.User) (int, error) {
	if user.Name == "" {
		return 0, fmt.Errorf("name cannot be empty")
	}
	if user.Username == "" {
		return 0, fmt.Errorf("username cannot be empty")
	}

	var userId int
	query := fmt.Sprintf("INSERT INTO %s (name, username, password_hash) VALUES ($1, $2, $3) RETURNING id", usersTable)

	err := s.db.Get(&userId, query, user.Name, user.Username, user.Password)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (r *AuthSQL) GetUser(username, password string) (todo.User, error) {
	var user todo.User
	query := fmt.Sprintf("SELECT * FROM %s WHERE username = $1 AND password_hash = $2", usersTable)
	err := r.db.Get(&user, query, username, password)

	return user, err
}

func (r *AuthSQL) UpdateUser(userId int, user todo.User) error {
	queryCheck := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE username = $1 AND id != $2", usersTable)
	var count int
	err := r.db.QueryRow(queryCheck, user.Username, userId).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check username uniqueness: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("username '%s' is already taken", user.Username)
	}

	queryUpdate := fmt.Sprintf("UPDATE %s SET name = COALESCE($1, name), username = COALESCE($2, username), password_hash = COALESCE($3, password_hash) WHERE id = $4", usersTable)
	_, err = r.db.Exec(queryUpdate, user.Name, user.Username, user.Password, userId)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}

func (r *AuthSQL) DeleteUser(userId int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	deleteTodoItemsQuery := `
		DELETE FROM todo_item 
		WHERE list_id IN (
			SELECT id FROM todo_list WHERE user_id = $1
		)
	`
	_, err = tx.Exec(deleteTodoItemsQuery, userId)
	if err != nil {
		return fmt.Errorf("failed to delete todo items: %v", err)
	}

	deleteTodoListsQuery := `
		DELETE FROM todo_list WHERE user_id = $1
	`
	_, err = tx.Exec(deleteTodoListsQuery, userId)
	if err != nil {
		return fmt.Errorf("failed to delete todo lists: %v", err)
	}

	deleteUserQuery := `
		DELETE FROM users WHERE id = $1
	`
	_, err = tx.Exec(deleteUserQuery, userId)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}
