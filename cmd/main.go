package main

import (
	"github.com/chmod-git/todo-app"
	"log"
)

func main() {
	server := new(todo.Server)

	if err := server.Run("8000"); err != nil {
		log.Fatalf("error occured while running the server: %v", err)
	}
}
