package main

import (
	"context"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/handler"
	"github.com/chmod-git/todo-app/pkg/repository"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	if err := initConfig(); err != nil {
		logrus.Fatalf("error initializing configs: %v", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("Error loading .env file: %v", err.Error())
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})

	if err != nil {
		logrus.Fatalf("error initializing DB: %v", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	server := new(todo.Server)

	go func() {
		if err := server.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running the server: %v", err)
		}
	}()

	logrus.Print("Todo-App Started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	logrus.Print("Todo-App Finished")

	if err = server.Shutdown(context.Background()); err != nil {
		logrus.Fatalf("error occured while shutting down the server: %v", err)
	}

	if err = db.Close(); err != nil {
		logrus.Fatalf("error occured while closing the database: %v", err)
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
