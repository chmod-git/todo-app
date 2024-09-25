package todo

import "fmt"

type TodoList struct {
	Id          int    `json:"id" db:"id"`
	UserId      int    `json:"-" db:"user_id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
}

type TodoItem struct {
	Id          int    `json:"id" db:"id"`
	ListId      int    `json:"-" db:"list_id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	Done        bool   `json:"done" db:"done"`
}

type UpdateListInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

func (i *UpdateListInput) Validate() error {
	if i.Description == nil && i.Title == nil {
		return fmt.Errorf("update structure has no values")
	}

	return nil
}

type UpdateItemInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Done        *bool   `json:"done"`
}

func (i *UpdateItemInput) Validate() error {
	if i.Description == nil && i.Title == nil && i.Done == nil {
		return fmt.Errorf("update structure has no values")
	}

	return nil
}
