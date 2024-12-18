package tg_bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

var mux sync.Mutex

func Account(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID = update.Message.Chat.ID
	if _, exists := accountData[chatID]; !exists || userToken == "" {
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

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Your account:\n Name: %s\n Username: %s\n Password: %s", userInfo.Name, userInfo.Username, userInfo.Password))
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func EditAcc(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	mu.Lock()
	accountData[chatID] = []string{}
	mu.Unlock()

	sendYesNoQuestion(bot, chatID, "Change your name:", "edit_name_yes", "edit_name_no")
}

func DeleteAcc(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID

	sendYesNoQuestion(bot, chatID, "Are you sure you want to delete your account?", "delete_acc_yes", "delete_acc_no")
}

func LogOut(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	userToken = ""
	sendMessage(bot, update.CallbackQuery.Message.Chat.ID, "You have been logged out.")
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
