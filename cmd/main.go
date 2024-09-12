package main

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/handler"
	"log"
)

func main() {
	router := new(handler.Handler)
	server := new(todo.Server)

	if err := server.Run("8000", router.InitRoutes()); err != nil {
		log.Fatalf("error occured while running the server: %v", err)
	}
}
