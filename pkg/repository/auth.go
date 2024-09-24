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
	query := fmt.Sprintf("SELECT id FROM %s WHERE username=$1 AND password_hash=$2", usersTable)
	err := r.db.Get(&user, query, username, password)

	return user, err
}
