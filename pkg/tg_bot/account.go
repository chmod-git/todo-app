package tg_bot

import (
	"fmt"
	"github.com/chmod-git/todo-app"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func Account(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	session := GetSession(chatID)

	if session.UserToken == "" {
		Auth(bot, update)
		return
	}

	editAccButton := tgbotapi.NewInlineKeyboardButtonData("Edit", "edit_acc")
	deleteAccButton := tgbotapi.NewInlineKeyboardButtonData("Delete", "delete_acc")
	logOutButton := tgbotapi.NewInlineKeyboardButtonData("Log out", "log_out")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(editAccButton),
		tgbotapi.NewInlineKeyboardRow(deleteAccButton),
		tgbotapi.NewInlineKeyboardRow(logOutButton),
	)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
		"Your account:\n Name: %s\n Username: %s\n Password: %s",
		session.UserInfo.Name, session.UserInfo.Username, session.UserInfo.Password,
	))
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func EditAcc(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	session := GetSession(chatID)

	session.AccountData = []string{}
	session.Status = "edit_name"

	SendYesNoQuestion(bot, chatID, "Change your name:", "edit_name_yes", "edit_name_no")
}

func DeleteAcc(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	SendYesNoQuestion(bot, chatID, "Are you sure you want to delete your account?", "delete_acc_yes", "delete_acc_no")
}

func LogOut(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	session := GetSession(chatID)

	session.UserToken = ""
	session.UserInfo = todo.User{}
	session.AccountData = []string{}
	session.Status = ""

	SendMessage(bot, chatID, "You have been logged out.")
}

func SendYesNoQuestion(bot *tgbotapi.BotAPI, chatID int64, question, yesCallback, noCallback string) {
	yesButton := tgbotapi.NewInlineKeyboardButtonData("Yes", yesCallback)
	noButton := tgbotapi.NewInlineKeyboardButtonData("No", noCallback)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(yesButton, noButton),
	)

	msg := tgbotapi.NewMessage(chatID, question)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func HandleEditAccountMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)
	text := update.Message.Text

	switch session.Status {
	case "edit_name":
		session.AccountData = append(session.AccountData, text)
		session.Status = "edit_username"
		SendYesNoQuestion(bot, update.Message.Chat.ID, "Change your username:", "edit_username_yes", "edit_username_no")
	case "edit_username":
		session.AccountData = append(session.AccountData, text)
		session.Status = "edit_password"
		SendYesNoQuestion(bot, update.Message.Chat.ID, "Change your password:", "edit_password_yes", "edit_password_no")
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
			SendMessage(bot, chatID, "This username is already taken. Try again.")
			session.AccountData = session.AccountData[:1]
			session.Status = "edit_username"
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to update: %v", string(response))
			SendMessage(bot, chatID, "Unknown error. Try again.")
			session.AccountData = session.AccountData[:1]
			session.Status = "edit_username"
			return
		}

		session.UserInfo = input
		SendMessage(bot, chatID, "Account was successfully updated.")
	default:
		SendMessage(bot, chatID, "Unknown operation.")
	}
}
