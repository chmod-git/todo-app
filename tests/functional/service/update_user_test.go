package service

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateUser(t *testing.T) {
	mockRepo := new(MockAuthorizationRepository)
	service := service.NewAuthService(mockRepo)

	userID := 1
	updatedUser := todo.User{Username: "updatedUser", Password: "newPassword"}
	mockRepo.On("UpdateUser", userID, updatedUser).Return(nil)

	err := service.UpdateUser(userID, updatedUser)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
