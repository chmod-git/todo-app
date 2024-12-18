package tg_bot

import (
	"encoding/json"
	"github.com/chmod-git/todo-app"
	my_lib "github.com/chmod-git/todo-app/pkg/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func Auth(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	signUpButton := tgbotapi.NewInlineKeyboardButtonData("Sign-Up", "sign_up")
	signInButton := tgbotapi.NewInlineKeyboardButtonData("Sign-In", "sign_in")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(signUpButton),
		tgbotapi.NewInlineKeyboardRow(signInButton),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "To start using me, please authorize:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func SignIn(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID

	session := GetSession(chatID)

	session.AccountData = []string{""}
	session.Status = "sign_in"

	SendMessage(bot, chatID, "Logging into existing account...")
	SendMessage(bot, chatID, "Enter your username:")
}

func SignUp(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID

	session := GetSession(chatID)

	session.AccountData = []string{}
	session.Status = "sign_up"

	SendMessage(bot, chatID, "Creating new account...")
	SendMessage(bot, chatID, "Create your name:")
}

func SendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		logrus.Errorf("Failed to send message to %d: %v", chatID, err)
	}
}

func HandleSignInMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)
	text := update.Message.Text

	switch len(session.AccountData) {
	case 1:
		session.AccountData = append(session.AccountData, text)
		SendMessage(bot, chatID, "Enter your password:")
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
			SendMessage(bot, chatID, "Wrong username or password. Try again.")
			Auth(bot, update)
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to sign in: %v", string(response))
			SendMessage(bot, chatID, "Wrong username or password. Try again.")
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

		SendMessage(bot, chatID, "Authorization successful.")
		SendMessage(bot, chatID, "Use these commands:\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		SendMessage(bot, chatID, "Error. Try again by clicking on Sign-In.")
	}
}

func HandleSignUpMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)
	text := update.Message.Text

	switch len(session.AccountData) {
	case 0:
		session.AccountData = append(session.AccountData, text)
		SendMessage(bot, chatID, "Create your username:")
	case 1:
		session.AccountData = append(session.AccountData, text)
		SendMessage(bot, chatID, "Create your password:")
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
			SendMessage(bot, chatID, "This username is already taken. Try again.")
			Auth(bot, update)
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to sign up: %v", string(response))
			SendMessage(bot, chatID, "Unknown error. Try again.")
			Auth(bot, update)
			return
		}

		SendMessage(bot, chatID, "New account was successfully created.")

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

		SendMessage(bot, chatID, "Use these commands:\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		SendMessage(bot, chatID, "Error. Try again by clicking on Sign-In.")
	}
}
