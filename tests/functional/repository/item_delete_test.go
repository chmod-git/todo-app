package repository

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chmod-git/todo-app/pkg/repository"
	"github.com/jmoiron/sqlx"
	"testing"
)

func TestItemSQL_Delete(t *testing.T) {
	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		userId int
		listId int
		itemId int
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErr        bool
		mockDB         func() *sqlx.DB
		mockCheckList  func(userId, listId int) (bool, error)
		mockExistsItem func(itemId, listId int) error
	}{
		{
			name: "successful deletion",
			fields: fields{
				db: nil,
			},
			args: args{
				userId: 1,
				listId: 10,
				itemId: 100,
			},
			wantErr: false,
			mockDB: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("DELETE FROM todo_items WHERE id = $1 AND list_id = $2").
					WithArgs(100, 10).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return sqlx.NewDb(db, "sqlmock")
			},
			mockCheckList: func(userId, listId int) (bool, error) {
				return true, nil
			},
			mockExistsItem: func(itemId, listId int) error {
				return nil
			},
		},
		{
			name: "list does not exist",
			fields: fields{
				db: nil,
			},
			args: args{
				userId: 1,
				listId: 10,
				itemId: 100,
			},
			wantErr: true,
			mockDB: func() *sqlx.DB {
				return nil
			},
			mockCheckList: func(userId, listId int) (bool, error) {
				return false, nil
			},
			mockExistsItem: func(itemId, listId int) error {
				return nil
			},
		},
		{
			name: "item does not exist",
			fields: fields{
				db: nil,
			},
			args: args{
				userId: 1,
				listId: 10,
				itemId: 100,
			},
			wantErr: true,
			mockDB: func() *sqlx.DB {
				return nil
			},
			mockCheckList: func(userId, listId int) (bool, error) {
				return true, nil
			},
			mockExistsItem: func(itemId, listId int) error {
				return fmt.Errorf("item does not exist")
			},
		},
		{
			name: "delete query error",
			fields: fields{
				db: nil,
			},
			args: args{
				userId: 1,
				listId: 10,
				itemId: 100,
			},
			wantErr: true,
			mockDB: func() *sqlx.DB {
				db, mock, _ := sqlmock.New()
				mock.ExpectExec("DELETE FROM todo_items WHERE id = \\$1 AND list_id = \\$2").
					WithArgs(100, 10).
					WillReturnError(fmt.Errorf("delete query error"))
				return sqlx.NewDb(db, "sqlmock")
			},
			mockCheckList: func(userId, listId int) (bool, error) {
				return true, nil
			},
			mockExistsItem: func(itemId, listId int) error {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockDB != nil {
				tt.fields.db = tt.mockDB()
			}
			s := repository.NewItemSQL(tt.fields.db)

			err := s.Delete(tt.args.userId, tt.args.listId, tt.args.itemId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
