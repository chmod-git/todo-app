package tg_bot

import (
	"fmt"
	"github.com/chmod-git/todo-app"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Account(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	session := getSession(chatID)

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
	session := getSession(chatID)

	session.AccountData = []string{}
	session.Status = "edit_name"

	sendYesNoQuestion(bot, chatID, "Change your name:", "edit_name_yes", "edit_name_no")
}

func DeleteAcc(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	sendYesNoQuestion(bot, chatID, "Are you sure you want to delete your account?", "delete_acc_yes", "delete_acc_no")
}

func LogOut(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	session := getSession(chatID)

	session.UserToken = ""
	session.UserInfo = todo.User{}
	session.AccountData = []string{}
	session.Status = ""

	sendMessage(bot, chatID, "You have been logged out.")
}

func sendYesNoQuestion(bot *tgbotapi.BotAPI, chatID int64, question, yesCallback, noCallback string) {
	yesButton := tgbotapi.NewInlineKeyboardButtonData("Yes", yesCallback)
	noButton := tgbotapi.NewInlineKeyboardButtonData("No", noCallback)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(yesButton, noButton),
	)

	msg := tgbotapi.NewMessage(chatID, question)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}
