package repository

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/repository"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestAuthSQL_CreateUser(t *testing.T) {
	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		user todo.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
		mockDB  func() *sqlx.DB
	}{
		{
			name: "successful user creation",
			fields: fields{
				db: nil,
			},
			args: args{
				user: todo.User{
					Name:     "John Doe",
					Username: "johndoe",
					Password: "hashedpassword123",
				},
			},
			want:    1,
			wantErr: false,
			mockDB: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("John Doe", "johndoe", "hashedpassword123").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				return sqlx.NewDb(db, "sqlmock")
			},
		},
		{
			name: "validation error - empty name",
			fields: fields{
				db: nil,
			},
			args: args{
				user: todo.User{
					Name:     "",
					Username: "johndoe",
					Password: "hashedpassword123",
				},
			},
			want:    0,
			wantErr: true,
			mockDB: func() *sqlx.DB {
				return nil
			},
		},
		{
			name: "validation error - empty username",
			fields: fields{
				db: nil,
			},
			args: args{
				user: todo.User{
					Name:     "John Doe",
					Username: "",
					Password: "hashedpassword123",
				},
			},
			want:    0,
			wantErr: true,
			mockDB: func() *sqlx.DB {
				return nil
			},
		},
		{
			name: "database error",
			fields: fields{
				db: nil,
			},
			args: args{
				user: todo.User{
					Name:     "John Doe",
					Username: "johndoe",
					Password: "hashedpassword123",
				},
			},
			want:    0,
			wantErr: true,
			mockDB: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("John Doe", "johndoe", "hashedpassword123").
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

			s := repository.NewAuthSQL(tt.fields.db)
			got, err := s.CreateUser(tt.args.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CreateUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}
