package tg_bot

import (
	"encoding/json"
	"github.com/chmod-git/todo-app"
	my_lib "github.com/chmod-git/todo-app/pkg/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"os"
)

type UserSession struct {
	UserInfo    todo.User
	UserToken   string
	AccountData []string
	Status      string
}

func getSession(chatID int64) *UserSession {
	if _, exists := userSessions[chatID]; !exists {
		userSessions[chatID] = &UserSession{
			AccountData: []string{},
			Status:      "",
		}
	}
	return userSessions[chatID]
}

var userSessions = make(map[int64]*UserSession)

func LaunchBot() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		logrus.Fatalf("Error occurred while creating the bot: %v", err)
	}

	bot.Debug = false
	logrus.Infof("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		go func(update tgbotapi.Update) {
			chatID := getChatID(update)
			if chatID == 0 {
				logrus.Warn("Could not determine chat ID")
				return
			}

			session := getSession(chatID)

			if update.Message != nil && update.Message.IsCommand() {
				handleCommand(bot, update)
			}

			if update.CallbackQuery != nil {
				handleCallback(bot, update)
			}

			if update.Message != nil && !update.Message.IsCommand() {
				switch session.Status {
				case "sign_in":
					handleSignInMessage(bot, update, session)
				case "sign_up":
					handleSignUpMessage(bot, update, session)
				case "edit_name", "edit_username", "edit_password":
					handleEditAccountMessage(bot, update, session)
				}
			}
		}(update)
	}
}

func getChatID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		sendMessage(bot, update.Message.Chat.ID, "Hi! I'm bot created to help you manage your tasks")
		Auth(bot, update)
	case "auth":
		Auth(bot, update)
	case "account":
		Account(bot, update)
	case "help":
		sendMessage(bot, update.Message.Chat.ID, "Use these commands:\n /auth - to authorize\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		sendMessage(bot, update.Message.Chat.ID, "Unknown command. Enter /help to see all available commands.")
	}
}

func handleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := getChatID(update)
	session := getSession(chatID)

	switch update.CallbackQuery.Data {
	case "sign_in":
		SignIn(bot, update)
		session.Status = "sign_in"
	case "sign_up":
		SignUp(bot, update)
		session.Status = "sign_up"
	case "edit_acc":
		EditAcc(bot, update)
	case "edit_name_yes":
		sendMessage(bot, chatID, "Enter new name:")
		session.Status = "edit_name"
	case "edit_name_no":
		session.AccountData = append(session.AccountData, session.UserInfo.Name)
		sendYesNoQuestion(bot, chatID, "Change your username:", "edit_username_yes", "edit_username_no")
	case "edit_username_yes":
		sendMessage(bot, chatID, "Enter new username:")
		session.Status = "edit_username"
	case "edit_username_no":
		session.AccountData = append(session.AccountData, session.UserInfo.Username)
		sendYesNoQuestion(bot, chatID, "Change your password:", "edit_password_yes", "edit_password_no")
	case "edit_password_yes":
		sendMessage(bot, chatID, "Enter new password:")
		session.Status = "edit_password"
	case "edit_password_no":
		session.AccountData = append(session.AccountData, session.UserInfo.Password)
		session.Status = ""

		httpClient := NewHTTPClient("http://localhost:8000")

		headers := map[string]string{
			"Authorization": "Bearer " + session.UserToken,
		}
		input := todo.User{
			Name:     session.AccountData[0],
			Username: session.AccountData[1],
			Password: session.AccountData[2],
		}

		response, statusCode, err := httpClient.PUT("/auth/api/update", headers, input)
		if err != nil {
			logrus.Errorf("Failed to update: %v", err)
			sendMessage(bot, chatID, "This username is already taken. Try again.")
			session.AccountData = session.AccountData[:1]
			session.Status = "edit_username"
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to update: %v", string(response))
			sendMessage(bot, chatID, "Unknown error. Try again.")
			session.AccountData = session.AccountData[:1]
			session.Status = "edit_username"
			return
		}

		session.UserInfo = input
		sendMessage(bot, chatID, "Account was successfully updated.")
	case "delete_acc":
		DeleteAcc(bot, update)
	case "delete_acc_yes":
		httpClient := NewHTTPClient("http://localhost:8000")

		headers := map[string]string{
			"Authorization": "Bearer " + session.UserToken,
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
		session.UserToken = ""
		session.UserInfo = todo.User{}
		delete(userSessions, chatID)
	case "delete_acc_no":
		sendMessage(bot, chatID, "Operation was canceled.")
	case "log_out":
		LogOut(bot, update)
	}
}

func handleSignInMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := getChatID(update)
	text := update.Message.Text

	switch len(session.AccountData) {
	case 1:
		session.AccountData = append(session.AccountData, text)
		sendMessage(bot, chatID, "Enter your password:")
	case 2:
		session.AccountData = append(session.AccountData, text)
		logrus.Infof("User input: Username: %s, Password: %s", session.AccountData[1], session.AccountData[2])

		httpClient := NewHTTPClient("http://localhost:8000")
		input := my_lib.SignInInput{
			Username: session.AccountData[1],
			Password: session.AccountData[2],
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

		session.UserToken = signInResponse.Token

		headers := map[string]string{
			"Authorization": "Bearer " + session.UserToken,
		}
		response, _, _ = httpClient.POST("/auth/api/info", headers, input)
		var user todo.User
		err = json.Unmarshal(response, &user)
		session.AccountData[0] = user.Name
		session.UserInfo = todo.User{
			Name:     session.AccountData[0],
			Username: session.AccountData[1],
			Password: session.AccountData[2],
		}

		sendMessage(bot, chatID, "Authorization successful.")
		sendMessage(bot, chatID, "Use these commands:\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		sendMessage(bot, chatID, "Error. Try again by clicking on Sign-In.")
	}
}

func handleSignUpMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := getChatID(update)
	text := update.Message.Text

	switch len(session.AccountData) {
	case 0:
		session.AccountData = append(session.AccountData, text)
		sendMessage(bot, chatID, "Create your username:")
	case 1:
		session.AccountData = append(session.AccountData, text)
		sendMessage(bot, chatID, "Create your password:")
	case 2:
		session.AccountData = append(session.AccountData, text)
		logrus.Infof("User %d input: Name: %s, Username: %s, Password: %s",
			chatID, session.AccountData[0], session.AccountData[1], session.AccountData[2])

		httpClient := NewHTTPClient("http://localhost:8000")
		input := todo.User{
			Name:     session.AccountData[0],
			Username: session.AccountData[1],
			Password: session.AccountData[2],
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
			Username: session.AccountData[1],
			Password: session.AccountData[2],
		}

		response, _, _ = httpClient.POST("/auth/sign-in", nil, signInInput)
		var signInResponse AuthResponse
		json.Unmarshal(response, &signInResponse)

		session.UserToken = signInResponse.Token
		session.UserInfo = todo.User{
			Name:     session.AccountData[0],
			Username: session.AccountData[1],
			Password: session.AccountData[2],
		}

		sendMessage(bot, chatID, "Use these commands:\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		sendMessage(bot, chatID, "Error. Try again by clicking on Sign-In.")
	}
}

func handleEditAccountMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := getChatID(update)
	text := update.Message.Text

	switch session.Status {
	case "edit_name":
		session.AccountData = append(session.AccountData, text)
		session.Status = "edit_username"
		sendYesNoQuestion(bot, update.Message.Chat.ID, "Change your username:", "edit_username_yes", "edit_username_no")
	case "edit_username":
		session.AccountData = append(session.AccountData, text)
		session.Status = "edit_password"
		sendYesNoQuestion(bot, update.Message.Chat.ID, "Change your password:", "edit_password_yes", "edit_password_no")
	case "edit_password":
		session.AccountData = append(session.AccountData, text)
		session.Status = ""

		httpClient := NewHTTPClient("http://localhost:8000")

		headers := map[string]string{
			"Authorization": "Bearer " + session.UserToken,
		}
		input := todo.User{
			Name:     session.AccountData[0],
			Username: session.AccountData[1],
			Password: session.AccountData[2],
		}

		response, statusCode, err := httpClient.PUT("/auth/api/update", headers, input)
		if err != nil {
			logrus.Errorf("Failed to update: %v", err)
			sendMessage(bot, chatID, "This username is already taken. Try again.")
			session.AccountData = session.AccountData[:1]
			session.Status = "edit_username"
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to update: %v", string(response))
			sendMessage(bot, chatID, "Unknown error. Try again.")
			session.AccountData = session.AccountData[:1]
			session.Status = "edit_username"
			return
		}

		session.UserInfo = input
		sendMessage(bot, chatID, "Account was successfully updated.")
	default:
		sendMessage(bot, chatID, "Unknown operation.")
	}
}
