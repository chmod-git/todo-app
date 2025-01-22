package repository

import (
	"github.com/chmod-git/todo-app"
	"github.com/stretchr/testify/assert"
	sqlmock "github.com/zhashkevych/go-sqlxmock"
	"log"
	"testing"
)

func TestTodoItem_Create(t *testing.T) {
	db, mock, err := sqlmock.Newx()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := NewItemSQL(db)

	type args struct {
		userId int
		item   todo.TodoItem
	}
	type mockBehaviour func(args args, id int)

	tests := []struct {
		name          string
		input         args
		mockBehaviour mockBehaviour
		wantId        int
		wantErr       bool
	}{
		{
			name: "OK",
			input: args{
				userId: 1,
				item: todo.TodoItem{
					ListId:      1,
					Title:       "test title",
					Description: "test description",
					Done:        false,
				},
			},
			mockBehaviour: func(args args, id int) {
				existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
				mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM todo_list WHERE id = \$1 AND user_id = \$2\)`).
					WithArgs(args.item.ListId, args.userId).
					WillReturnRows(existsRows)

				itemRows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO todo_item").
					WithArgs(args.item.Title, args.item.Description, args.item.Done, args.item.ListId).
					WillReturnRows(itemRows)
			},
			wantId:  1,
			wantErr: false,
		},
		{
			name: "Empty fields",
			input: args{
				item: todo.TodoItem{
					ListId:      1,
					Title:       "test title",
					Description: "test description",
					Done:        false,
				},
			},
			mockBehaviour: func(args args, id int) {
				existsRows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
				mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM todo_list WHERE id = \$1 AND user_id = \$2\)`).
					WithArgs(args.item.ListId, args.userId).
					WillReturnRows(existsRows)
			},
			wantId:  1,
			wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehaviour(testCase.input, testCase.wantId)

			got, err := r.Create(testCase.input.userId, testCase.input.item.ListId, testCase.input.item)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.wantId, got)
			}
		})
	}
}
