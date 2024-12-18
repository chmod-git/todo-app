package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"sync"
)

var mu sync.Mutex

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

	sendMessage(bot, chatID, "Logging into existing account...")
	sendMessage(bot, chatID, "Enter your username:")
	mu.Lock()
	accountData[chatID] = []string{""}
	mu.Unlock()
}

func SignUp(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID

	sendMessage(bot, chatID, "Creating new account...")
	sendMessage(bot, chatID, "Create your name:")
	mu.Lock()
	accountData[chatID] = []string{}
	mu.Unlock()
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		logrus.Errorf("Failed to send message to %d: %v", chatID, err)
	}
}
