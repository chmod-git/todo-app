package main

import (
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/handler"
	"github.com/chmod-git/todo-app/pkg/repository"
	"github.com/chmod-git/todo-app/pkg/service"
	"log"
)

func main() {
	repos := repository.NewRepository()
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	server := new(todo.Server)

	if err := server.Run("8000", handlers.InitRoutes()); err != nil {
		log.Fatalf("error occured while running the server: %v", err)
	}
}
