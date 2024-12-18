package tg_bot

import (
	"encoding/json"
	"github.com/chmod-git/todo-app"
	my_lib "github.com/chmod-git/todo-app/pkg/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"os"
)

var messageStatus string
var chatID int64
var userInfo todo.User
var userToken string
var accountData = make(map[int64][]string)

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
			switch messageStatus {
			case "sign_in":
				handleSignInMessage(bot, update)
			case "sign_up":
				handleSignUpMessage(bot, update)
			case "edit_name", "edit_username", "edit_password":
				handleEditAccountMessage(bot, update)
			}
		}
	}
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		sendMessage(bot, update.Message.Chat.ID, "Hi! I'm bot created to help you manage your tasks")
		Auth(bot, update)
	case "account":
		Account(bot, update)
	case "help":
		sendMessage(bot, chatID, "Use these commands:\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		sendMessage(bot, update.Message.Chat.ID, "Unknown command. Enter /help to see all available commands.")
	}
}

func handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.CallbackQuery.Data {
	case "sign_in":
		SignIn(bot, update)
		messageStatus = "sign_in"
	case "sign_up":
		SignUp(bot, update)
		messageStatus = "sign_up"
	case "edit_acc":
		EditAcc(bot, update)
	case "edit_name_yes":
		sendMessage(bot, chatID, "Enter new name:")
		messageStatus = "edit_name"
	case "edit_name_no":
		accountData[chatID] = append(accountData[chatID], userInfo.Name)
		sendYesNoQuestion(bot, chatID, "Change your username:", "edit_username_yes", "edit_username_no")
	case "edit_username_yes":
		sendMessage(bot, chatID, "Enter new username:")
		messageStatus = "edit_username"
	case "edit_username_no":
		accountData[chatID] = append(accountData[chatID], userInfo.Username)
		sendYesNoQuestion(bot, chatID, "Change your password:", "edit_password_yes", "edit_password_no")
	case "edit_password_yes":
		sendMessage(bot, chatID, "Enter new password:")
		messageStatus = "edit_password"
	case "edit_password_no":
		accountData[chatID] = append(accountData[chatID], userInfo.Password)
		messageStatus = ""

		httpClient := NewHTTPClient("http://localhost:8000")

		headers := map[string]string{
			"Authorization": "Bearer " + userToken,
		}
		input := todo.User{
			Name:     accountData[chatID][0],
			Username: accountData[chatID][1],
			Password: accountData[chatID][2],
		}

		response, statusCode, err := httpClient.PUT("/auth/api/update", headers, input)
		if err != nil {
			logrus.Errorf("Failed to update: %v", err)
			sendMessage(bot, chatID, "This username is already taken. Try again.")
			accountData[chatID] = accountData[chatID][:1]
			messageStatus = "edit_username"
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to update: %v", string(response))
			sendMessage(bot, chatID, "Unknown error. Try again.")
			accountData[chatID] = accountData[chatID][:1]
			messageStatus = "edit_username"
			return
		}

		userInfo = input
		sendMessage(bot, chatID, "Account was successfully updated.")
	case "delete_acc":
		DeleteAcc(bot, update)
	case "delete_acc_yes":
		httpClient := NewHTTPClient("http://localhost:8000")

		headers := map[string]string{
			"Authorization": "Bearer " + userToken,
		}

		response, statusCode, err := httpClient.DELETE("/auth/api/delete", headers)
		if err != nil {
			logrus.Errorf("Failed to delete: %v", err)
			sendMessage(bot, chatID, "Unable to delete. Try again.")
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to delete: %v", string(response))
			sendMessage(bot, chatID, "Unknown error. Try again.")
			return
		}

		sendMessage(bot, chatID, "Account was deleted successfully.")
		userToken = ""
		userInfo = todo.User{}
		delete(accountData, chatID)
	case "delete_acc_no":
		sendMessage(bot, chatID, "Operation was canceled.")
	case "log_out":
		LogOut(bot, update)
	}
}

func handleSignInMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID = update.Message.Chat.ID
	text := update.Message.Text

	mu.Lock()
	defer mu.Unlock()

	if _, exists := accountData[chatID]; !exists {
		return
	}

	switch len(accountData[chatID]) {
	case 1:
		accountData[chatID] = append(accountData[chatID], text)
		sendMessage(bot, chatID, "Enter your password:")
	case 2:
		accountData[chatID] = append(accountData[chatID], text)
		logrus.Infof("User %d input: Username: %s, Password: %s",
			chatID, accountData[chatID][1], accountData[chatID][2])

		httpClient := NewHTTPClient("http://localhost:8000")
		input := my_lib.SignInInput{
			Username: accountData[chatID][1],
			Password: accountData[chatID][2],
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

		userToken = signInResponse.Token

		headers := map[string]string{
			"Authorization": "Bearer " + userToken,
		}
		response, _, _ = httpClient.POST("/auth/api/info", headers, input)
		var user todo.User
		err = json.Unmarshal(response, &user)
		accountData[chatID][0] = user.Name
		userInfo = todo.User{
			Name:     accountData[chatID][0],
			Username: accountData[chatID][1],
			Password: accountData[chatID][2],
		}

		sendMessage(bot, chatID, "Authorization successful.")
		sendMessage(bot, chatID, "Use these commands:\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		sendMessage(bot, chatID, "Error. Try again by clicking on Sign-In.")
	}
}

func handleSignUpMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID = update.Message.Chat.ID
	text := update.Message.Text

	mu.Lock()
	defer mu.Unlock()

	if _, exists := accountData[chatID]; !exists {
		return
	}

	switch len(accountData[chatID]) {
	case 0:
		accountData[chatID] = append(accountData[chatID], text)
		sendMessage(bot, chatID, "Create your username:")
	case 1:
		accountData[chatID] = append(accountData[chatID], text)
		sendMessage(bot, chatID, "Create your password:")
	case 2:
		accountData[chatID] = append(accountData[chatID], text)
		logrus.Infof("User %d input: Name: %s, Username: %s, Password: %s",
			chatID, accountData[chatID][0], accountData[chatID][1], accountData[chatID][2])

		httpClient := NewHTTPClient("http://localhost:8000")
		input := todo.User{
			Name:     accountData[chatID][0],
			Username: accountData[chatID][1],
			Password: accountData[chatID][2],
		}

		response, statusCode, err := httpClient.POST("/auth/sign-up", nil, input)
		if err != nil {
			logrus.Errorf("Failed to sign up: %v", err)
			sendMessage(bot, chatID, "This username is already taken. Try again.")
			Auth(bot, update)
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to sign up: %v", string(response))
			sendMessage(bot, chatID, "Unknown error. Try again.")
			Auth(bot, update)
			return
		}

		sendMessage(bot, chatID, "New account was successfully created.")

		signInInput := my_lib.SignInInput{
			Username: accountData[chatID][1],
			Password: accountData[chatID][2],
		}

		response, _, _ = httpClient.POST("/auth/sign-in", nil, signInInput)
		var signInResponse AuthResponse
		json.Unmarshal(response, &signInResponse)

		userToken = signInResponse.Token
		userInfo = todo.User{
			Name:     accountData[chatID][0],
			Username: accountData[chatID][1],
			Password: accountData[chatID][2],
		}

		sendMessage(bot, chatID, "Use these commands:\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		sendMessage(bot, chatID, "Error. Try again by clicking on Sign-In.")
	}
}

func handleEditAccountMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID = update.Message.Chat.ID
	text := update.Message.Text

	mux.Lock()
	defer mux.Unlock()

	if _, exists := accountData[chatID]; !exists {
		return
	}

	switch messageStatus {
	case "edit_name":
		accountData[chatID] = append(accountData[chatID], text)
		messageStatus = "edit_username"
		sendYesNoQuestion(bot, update.Message.Chat.ID, "Change your username:", "edit_username_yes", "edit_username_no")
	case "edit_username":
		accountData[chatID] = append(accountData[chatID], text)
		messageStatus = "edit_password"
		sendYesNoQuestion(bot, update.Message.Chat.ID, "Change your password:", "edit_password_yes", "edit_password_no")
	case "edit_password":
		accountData[chatID] = append(accountData[chatID], text)
		messageStatus = ""

		httpClient := NewHTTPClient("http://localhost:8000")

		headers := map[string]string{
			"Authorization": "Bearer " + userToken,
		}
		input := todo.User{
			Name:     accountData[chatID][0],
			Username: accountData[chatID][1],
			Password: accountData[chatID][2],
		}

		response, statusCode, err := httpClient.PUT("/auth/api/update", headers, input)
		if err != nil {
			logrus.Errorf("Failed to update: %v", err)
			sendMessage(bot, chatID, "This username is already taken. Try again.")
			accountData[chatID] = accountData[chatID][:1]
			messageStatus = "edit_username"
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to update: %v", string(response))
			sendMessage(bot, chatID, "Unknown error. Try again.")
			accountData[chatID] = accountData[chatID][:1]
			messageStatus = "edit_username"
			return
		}

		userInfo = input
		sendMessage(bot, chatID, "Account was successfully updated.")
	default:
		sendMessage(bot, chatID, "Unknown operation.")
	}
}
