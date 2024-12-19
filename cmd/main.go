package main

import (
	"context"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/handler"
	"github.com/chmod-git/todo-app/pkg/repository"
	"github.com/chmod-git/todo-app/pkg/service"
	"github.com/chmod-git/todo-app/pkg/tg_bot"
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

	postgres, err := repository.NewPostgresDB(repository.PostgresConfig{
		Host:     viper.GetString("postgres.host"),
		Port:     viper.GetString("postgres.port"),
		Username: viper.GetString("postgres.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("postgres.dbname"),
		SSLMode:  viper.GetString("postgres.sslmode"),
	})

	if err != nil {
		logrus.Fatalf("error initializing DB: %v", err.Error())
	}

	repos := repository.NewRepository(postgres)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	server := new(todo.Server)

	go func() {
		if err := server.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running the server: %v", err)
		}
	}()

	logrus.Print("Todo-App Started")

	auth := tg_bot.NewAuthorizationService()
	redis := repository.NewRedisRepository(
		viper.GetInt("redis.databases.lists"),
		viper.GetInt("redis.databases.items"),
		viper.GetString("redis.host"),
		viper.GetString("redis.port"),
		os.Getenv("DB_PASSWORD"),
	)

	bot := tg_bot.NewTelegramBotService(tg_bot.NewAccountService(auth), tg_bot.NewAuthorizationService(), tg_bot.NewListService(auth, redis), tg_bot.NewTaskService(auth, redis))

	go bot.LaunchBot()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit

	logrus.Print("Todo-App Finished")

	if err = server.Shutdown(context.Background()); err != nil {
		logrus.Fatalf("error occured while shutting down the server: %v", err)
	}

	if err = postgres.Close(); err != nil {
		logrus.Fatalf("error occured while closing the database: %v", err)
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
