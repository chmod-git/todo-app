package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Account interface {
	ManageAccount(bot *tgbotapi.BotAPI, update tgbotapi.Update)
	EditAcc(bot *tgbotapi.BotAPI, update tgbotapi.Update)
	HandleEditAccountMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession)
	DeleteAcc(bot *tgbotapi.BotAPI, update tgbotapi.Update)
	LogOut(bot *tgbotapi.BotAPI, update tgbotapi.Update)
}

type Authorization interface {
	Auth(bot *tgbotapi.BotAPI, update tgbotapi.Update)
	SignIn(bot *tgbotapi.BotAPI, update tgbotapi.Update)
	HandleSignInMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession)
	SignUp(bot *tgbotapi.BotAPI, update tgbotapi.Update)
	HandleSignUpMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession)
}

type List interface {
	Lists(bot *tgbotapi.BotAPI, update tgbotapi.Update)
	ManageLists(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession)
	PromptListSelection(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action string)
	ManageListInfo(bot *tgbotapi.BotAPI, chatID int64, listID string, session *UserSession)
	HandleManageListInfoMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession)
	DeleteList(bot *tgbotapi.BotAPI, chatID int64, listID string, session *UserSession)
	AddList(bot *tgbotapi.BotAPI, chatID int64, session *UserSession)
	HandleAddListMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession)
}

type Task interface {
	EditListTasks(bot *tgbotapi.BotAPI, chatID int64, listId string, session *UserSession)
	PromptTaskSelection(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action string)
	HandleTaskAction(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action, taskID string)
	AddTask(bot *tgbotapi.BotAPI, chatID int64, session *UserSession)
	HandleAddTaskMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession)
	UpdateTask(bot *tgbotapi.BotAPI, chatID int64, taskID string, session *UserSession)
	HandleUpdateTaskMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession)
	DeleteTask(bot *tgbotapi.BotAPI, chatID int64, taskID string, session *UserSession)
}

type TelegramBot interface {
	LaunchBot()
	HandleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update)
	HandleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update)
}
