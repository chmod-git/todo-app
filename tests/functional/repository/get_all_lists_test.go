package repository

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/repository"
	"github.com/jmoiron/sqlx"
	"reflect"
	"testing"
)

func TestListSQL_GetAll(t *testing.T) {
	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		userId int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []todo.TodoList
		wantErr bool
		mockDB  func() *sqlx.DB
	}{
		{
			name: "successful retrieval",
			fields: fields{
				db: nil,
			},
			args: args{
				userId: 1,
			},
			want: []todo.TodoList{
				{Id: 1, Title: "List 1", Description: "Description 1"},
				{Id: 2, Title: "List 2", Description: "Description 2"},
			},
			wantErr: false,
			mockDB: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM todo_lists tl WHERE tl.user_id = $1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description"}).
						AddRow(1, "List 1", "Description 1").
						AddRow(2, "List 2", "Description 2"))
				return sqlx.NewDb(db, "sqlmock")
			},
		},
		{
			name: "no results",
			fields: fields{
				db: nil,
			},
			args: args{
				userId: 2,
			},
			want:    []todo.TodoList{},
			wantErr: false,
			mockDB: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM todo_lists tl WHERE tl.user_id = $1").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description"}))
				return sqlx.NewDb(db, "sqlmock")
			},
		},
		{
			name: "database error",
			fields: fields{
				db: nil,
			},
			args: args{
				userId: 3,
			},
			want:    nil,
			wantErr: true,
			mockDB: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("SELECT tl.id, tl.title, tl.description FROM todo_lists tl WHERE tl.user_id = $1").
					WithArgs(3).
					WillReturnError(fmt.Errorf("database error"))
				return sqlx.NewDb(db, "sqlmock")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockDB != nil {
				tt.fields.db = tt.mockDB()
			}

			s := repository.NewListSQL(tt.fields.db)
			got, err := s.GetAll(tt.args.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAll() got = %v, want %v", got, tt.want)
			}
		})
	}
}
