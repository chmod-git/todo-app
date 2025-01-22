package handler

import (
	"errors"
	"fmt"
	"github.com/chmod-git/todo-app/pkg/service"
	mock_service "github.com/chmod-git/todo-app/pkg/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_userIdentity(t *testing.T) {
	type mockBehaviour func(s *mock_service.MockAuthorization, token string)
	testTable := []struct {
		name                string
		headerName          string
		headerValue         string
		token               string
		mockBehaviour       mockBehaviour
		expectedStatusCode  int
		expectedRequestBody string
	}{
		{
			name:        "OK",
			headerName:  "Authorization",
			headerValue: "Bearer j3287hygf7egwt23t4uy",
			token:       "j3287hygf7egwt23t4uy",
			mockBehaviour: func(s *mock_service.MockAuthorization, token string) {
				s.EXPECT().ParseToken(token).Return(1, nil)
			},
			expectedStatusCode:  200,
			expectedRequestBody: "1",
		},
		{
			name:        "No header",
			headerName:  "Authorization",
			headerValue: "",
			token:       "j3287hygf7egwt23t4uy",
			mockBehaviour: func(s *mock_service.MockAuthorization, token string) {
			},
			expectedStatusCode:  401,
			expectedRequestBody: `{"message":"empty authorization header"}`,
		},
		{
			name:        "Invalid header",
			headerName:  "Authorization",
			headerValue: "Berer j3287hygf7egwt23t4uy",
			token:       "j3287hygf7egwt23t4uy",
			mockBehaviour: func(s *mock_service.MockAuthorization, token string) {
			},
			expectedStatusCode:  401,
			expectedRequestBody: `{"message":"invalid authorization header"}`,
		},
		{
			name:        "Empty token",
			headerName:  "Authorization",
			headerValue: "Bearer ",
			token:       "",
			mockBehaviour: func(s *mock_service.MockAuthorization, token string) {
			},
			expectedStatusCode:  401,
			expectedRequestBody: `{"message":"token is empty"}`,
		},
		{
			name:        "Internal server error",
			headerName:  "Authorization",
			headerValue: "Bearer j3287hygf7egwt23t4uy",
			token:       "j3287hygf7egwt23t4uy",
			mockBehaviour: func(s *mock_service.MockAuthorization, token string) {
				s.EXPECT().ParseToken(token).Return(1, errors.New("internal server error"))
			},
			expectedStatusCode:  401,
			expectedRequestBody: `{"message":"invalid authorization header"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			auth := mock_service.NewMockAuthorization(c)
			testCase.mockBehaviour(auth, testCase.token)

			services := &service.Service{Authorization: auth}
			handler := NewHandler(services)

			r := gin.New()
			r.GET("/protected", handler.userIdentity, func(c *gin.Context) {
				id, _ := c.Get(userCtx)
				c.String(200, fmt.Sprintf("%d", id.(int)))
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", testCase.headerValue)

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.expectedStatusCode, w.Code)
			assert.Equal(t, testCase.expectedRequestBody, w.Body.String())
		})
	}
}
