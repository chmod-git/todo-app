package handler

import (
	"bytes"
	"errors"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/service"
	mock_service "github.com/chmod-git/todo-app/pkg/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_SignUp(t *testing.T) {
	type mockBehaviour func(s *mock_service.MockAuthorization, user todo.User)

	testTable := []struct {
		name                string
		inputBody           string
		inputUser           todo.User
		mockBehaviour       mockBehaviour
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"name": "Andy", "username": "/|ndy", "password": "qwerty"}`,
			inputUser: todo.User{
				Name:     "Andy",
				Username: "/|ndy",
				Password: "qwerty",
			},
			mockBehaviour: func(s *mock_service.MockAuthorization, user todo.User) {
				s.EXPECT().CreateUser(user).Return(1, nil)
			},
			expectedStatusCode:  200,
			expectedRequestBody: `{"id":1}`,
		},
		{
			name:                "Empty Fields",
			inputBody:           `{"username": "empty", "password": "qwerty"}`,
			mockBehaviour:       func(s *mock_service.MockAuthorization, user todo.User) {},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Internal server error",
			inputBody: `{"name": "Andy", "username": "/|ndy", "password": "qwerty"}`,
			inputUser: todo.User{
				Name:     "Andy",
				Username: "/|ndy",
				Password: "qwerty",
			},
			mockBehaviour: func(s *mock_service.MockAuthorization, user todo.User) {
				s.EXPECT().CreateUser(user).Return(0, errors.New("internal server error"))
			},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"internal server error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehaviour(auth, testCase.inputUser)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			r := gin.New()
			r.POST("/sign-up", handler.SignUp)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/sign-up",
				bytes.NewBufferString(testCase.inputBody))

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}

func TestHandler_updateUser(t *testing.T) {
	type mockBehaviour func(s *mock_service.MockAuthorization, userId int, user todo.User, token string)
	testTable := []struct {
		name                string
		inputBody           string
		inputUser           todo.User
		token               string
		inputUserId         int
		mockBehaviour       mockBehaviour
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:      "OK",
			inputBody: `{"name": "Andy", "username": "/|ndy", "password": "qwerty"}`,
			inputUser: todo.User{
				Name:     "Andy",
				Username: "/|ndy",
				Password: "qwerty",
			},
			token:       "j3287hygf7egwt23t4uy",
			inputUserId: 1,
			mockBehaviour: func(s *mock_service.MockAuthorization, userId int, user todo.User, token string) {
				s.EXPECT().UpdateUser(userId, user).Return(nil)
				s.EXPECT().ParseToken(token).Return(1, nil)
			},
			expectedStatusCode:  200,
			expectedRequestBody: `{"status":"ok"}`,
		},
		{
			name:      "Wrong token",
			inputBody: `{"name": "Andy", "username": "/|ndy", "password": "qwerty"}`,
			inputUser: todo.User{
				Name:     "Andy",
				Username: "/|ndy",
				Password: "qwerty",
			},
			token:       "j3287hygf7e4uy",
			inputUserId: 0,
			mockBehaviour: func(s *mock_service.MockAuthorization, userId int, user todo.User, token string) {
				s.EXPECT().ParseToken(token).Return(0, errors.New("wrong token"))
			},
			expectedStatusCode:  401,
			expectedRequestBody: `{"message":"invalid authorization header"}`,
		},
		{
			name:      "Wrong user input",
			inputBody: `{"name": "Andy", "password": "qwerty"}`,
			inputUser: todo.User{
				Name:     "Andy",
				Username: "/|ndy",
				Password: "qwerty",
			},
			token:       "j3287hygf7e4uy",
			inputUserId: 0,
			mockBehaviour: func(s *mock_service.MockAuthorization, userId int, user todo.User, token string) {
				s.EXPECT().ParseToken(token).Return(1, nil)
			},
			expectedStatusCode:  400,
			expectedRequestBody: `{"message":"invalid user input"}`,
		},
		{
			name:      "Internal server error",
			inputBody: `{"name": "Andy", "username": "/|ndy", "password": "qwerty"}`,
			inputUser: todo.User{
				Name:     "Andy",
				Username: "/|ndy",
				Password: "qwerty",
			},
			token:       "j3287hygf7e4uy",
			inputUserId: 0,
			mockBehaviour: func(s *mock_service.MockAuthorization, userId int, user todo.User, token string) {
				s.EXPECT().UpdateUser(userId, user).Return(errors.New("internal server error"))
				s.EXPECT().ParseToken(token).Return(0, nil)
			},
			expectedStatusCode:  500,
			expectedRequestBody: `{"message":"internal server error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehaviour(auth, testCase.inputUserId, testCase.inputUser, testCase.token)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			r := gin.New()
			api := r.Group("", handler.userIdentity)
			{
				api.POST("/update", handler.updateUser)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/update",
				bytes.NewBufferString(testCase.inputBody))
			req.Header.Set("Authorization", "Bearer "+testCase.token)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}
