package handler

import (
	"bytes"
	"encoding/json"
	"github.com/chmod-git/todo-app/pkg/handler"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignInHandler(t *testing.T) {
	mockService := new(MockAuthorizationService)
	handl := handler.NewHandler(&service.Service{mockService, nil, nil})

	router := gin.Default()
	router.POST("/auth/sign-in", handl.SignIn)

	input := handler.SignInInput{Username: "testuser", Password: "password"}
	token := "mocked_token"
	mockService.On("GenerateToken", input.Username, input.Password).Return(token, nil)

	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, token, response["token"])
	mockService.AssertExpectations(t)
}
