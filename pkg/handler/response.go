package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

func newErrorResponse(c *gin.Context, statusCode int, err string) {
	logrus.Errorf(err)
	c.AbortWithStatusJSON(statusCode, ErrorResponse{err})
}
