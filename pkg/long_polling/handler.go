package long_polling

import (
	"encoding/json"
	"github.com/chmod-git/todo-app"
	my_lib "github.com/chmod-git/todo-app/pkg/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"os"
)

var messageStatus string
var userTokens = make(map[string]string)
var userData = make(map[int64][]string)

func LaunchBot() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		logrus.Fatalf("error occured while creating the bot: %v", err)
	}

	bot.Debug = false
	logrus.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			handleCommand(bot, update)
		}

		if update.CallbackQuery != nil {
			handleCallback(bot, update)
		}

		if update.Message != nil && !update.Message.IsCommand() {
			if messageStatus == "signin" {
				handleSignInMessage(bot, update)
			} else if messageStatus == "signup" {
				handleSignUpMessage(bot, update)
			}
		}
	}
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		Auth(bot, update)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.CallbackQuery.Data {
	case "signin":
		SignIn(bot, update)
		messageStatus = "signin"
	case "signup":
		SignUp(bot, update)
		messageStatus = "signup"
	}
}

func handleSignInMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	mu.Lock()
	defer mu.Unlock()

	if _, exists := userData[chatID]; !exists {
		return
	}

	switch len(userData[chatID]) {
	case 0:
		userData[chatID] = append(userData[chatID], text)
		sendMessage(bot, chatID, "Enter your password:")
	case 1:
		userData[chatID] = append(userData[chatID], text)
		logrus.Infof("User %d input: Username: %s, Password: %s",
			chatID, userData[chatID][0], userData[chatID][1])

		httpClient := NewHTTPClient("http://localhost:8000")
		input := my_lib.SignInInput{
			Username: userData[chatID][0],
			Password: userData[chatID][1],
		}

		response, statusCode, err := httpClient.POST("/auth/sign-in", nil, input)
		if err != nil {
			logrus.Errorf("Failed to sign in: %v", err)
			sendMessage(bot, chatID, "Wrong username or password. Try again.")
			Auth(bot, update)
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to sign in: %v", string(response))
			sendMessage(bot, chatID, "Wrong username or password. Try again.")
			Auth(bot, update)
			return
		}

		var signInResponse AuthResponse
		err = json.Unmarshal(response, &signInResponse)
		if err != nil {
			logrus.Errorf("Failed to unmarshal response: %v", err)
		}

		userTokens[userData[chatID][0]] = signInResponse.Token
		sendMessage(bot, chatID, "Authorization successful.")

		delete(userData, chatID)
	default:
		sendMessage(bot, chatID, "Error. Try again by clicking on Sign-In.")
		delete(userData, chatID)
	}
}

func handleSignUpMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	mu.Lock()
	defer mu.Unlock()

	if _, exists := userData[chatID]; !exists {
		return
	}

	switch len(userData[chatID]) {
	case 0:
		userData[chatID] = append(userData[chatID], text)
		sendMessage(bot, chatID, "Create your username:")
	case 1:
		userData[chatID] = append(userData[chatID], text)
		sendMessage(bot, chatID, "Create your password:")
	case 2:
		userData[chatID] = append(userData[chatID], text)
		logrus.Infof("User %d input: Name: %s, Username: %s, Password: %s",
			chatID, userData[chatID][0], userData[chatID][1], userData[chatID][2])

		httpClient := NewHTTPClient("http://localhost:8000")
		input := todo.User{
			Name:     userData[chatID][0],
			Username: userData[chatID][1],
			Password: userData[chatID][2],
		}

		response, statusCode, err := httpClient.POST("/auth/sign-up", nil, input)
		if err != nil {
			logrus.Errorf("Failed to sign up: %v", err)
			sendMessage(bot, chatID, "This username is already taken. Try again.")
			Auth(bot, update)
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to sign up: %v", string(response))
			sendMessage(bot, chatID, "This username is already taken. Try again.")
			Auth(bot, update)
			return
		}

		sendMessage(bot, chatID, "New account was successfully created.")

		signInInput := my_lib.SignInInput{
			Username: userData[chatID][1],
			Password: userData[chatID][2],
		}

		response, _, _ = httpClient.POST("/auth/sign-in", nil, signInInput)
		var signInResponse AuthResponse
		json.Unmarshal(response, &signInResponse)
		userTokens[userData[chatID][1]] = signInResponse.Token

		delete(userData, chatID)
	default:
		sendMessage(bot, chatID, "Error. Try again by clicking on Sign-In.")
		delete(userData, chatID)
	}
}
