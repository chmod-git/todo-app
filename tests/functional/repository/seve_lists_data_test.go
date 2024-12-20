package repository

import (
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/repository"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"testing"
	"time"
)

func TestRedisRepository_SaveListsData(t *testing.T) {
	type fields struct {
		listClient *redis.Client
		itemClient *redis.Client
	}
	type args struct {
		chatID     string
		data       []todo.TodoList
		ttlSeconds int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		mockDB  func() *redis.Client
	}{
		{
			name: "successful save",
			fields: fields{
				listClient: nil,
				itemClient: nil,
			},
			args: args{
				chatID: "12345",
				data: []todo.TodoList{
					{Id: 1, Title: "List 1", Description: "Description 1"},
					{Id: 2, Title: "List 2", Description: "Description 2"},
				},
				ttlSeconds: 60,
			},
			wantErr: false,
			mockDB: func() *redis.Client {
				client, mock := redismock.NewClientMock()
				mock.ExpectSet("chat:12345", `[
					{"ID":1,"Title":"List 1","Description":"Description 1"},
					{"ID":2,"Title":"List 2","Description":"Description 2"}
				]`, 60*time.Second).SetVal("OK")
				return client
			},
		},
		{
			name: "serialization error",
			fields: fields{
				listClient: nil,
				itemClient: nil,
			},
			args: args{
				chatID:     "12345",
				data:       []todo.TodoList{},
				ttlSeconds: 60,
			},
			wantErr: true,
			mockDB: func() *redis.Client {
				return nil
			},
		},
		{
			name: "redis set error",
			fields: fields{
				listClient: nil,
				itemClient: nil,
			},
			args: args{
				chatID: "12345",
				data: []todo.TodoList{
					{Id: 1, Title: "List 1", Description: "Description 1"},
				},
				ttlSeconds: 60,
			},
			wantErr: true,
			mockDB: func() *redis.Client {
				client, mock := redismock.NewClientMock()
				mock.ExpectSet("chat:12345", `[
					{"ID":1,"Title":"List 1","Description":"Description 1"}
				]`, 60*time.Second).SetErr(fmt.Errorf("redis set error"))
				return client
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockDB != nil {
				tt.fields.listClient = tt.mockDB()
			}

			r := repository.RedisRepository{}
			err := r.SaveListsData(tt.args.chatID, tt.args.data, tt.args.ttlSeconds)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveListsData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
