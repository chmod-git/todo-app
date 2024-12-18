package tg_bot

import (
	"github.com/chmod-git/todo-app"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

type UserSession struct {
	UserInfo      todo.User
	CurrentListID string
	UserToken     string
	AccountData   []string
	ListData      todo.TodoList
	TaskData      todo.TodoItem
	Status        string
}

func GetSession(chatID int64) *UserSession {
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
			chatID := GetChatID(update)
			if chatID == 0 {
				logrus.Warn("Could not determine chat ID")
				return
			}

			session := GetSession(chatID)

			if update.Message != nil && update.Message.IsCommand() {
				HandleCommand(bot, update)
			}

			if update.CallbackQuery != nil {
				HandleCallback(bot, update)
			}

			if update.Message != nil && !update.Message.IsCommand() {
				switch session.Status {
				case "sign_in":
					HandleSignInMessage(bot, update, session)
				case "sign_up":
					HandleSignUpMessage(bot, update, session)
				case "edit_name", "edit_username", "edit_password":
					HandleEditAccountMessage(bot, update, session)
				case "add_task_title", "add_task_description":
					HandleAddTaskMessage(bot, update, session)
				case "edit_task_title", "edit_task_description":
					HandleUpdateTaskMessage(bot, update, session)
				}
			}
		}(update)
	}
}

func GetChatID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func HandleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		SendMessage(bot, update.Message.Chat.ID, "Hi! I'm bot created to help you manage your tasks")
		Auth(bot, update)
	case "auth":
		Auth(bot, update)
	case "account":
		Account(bot, update)
	case "lists":
		Lists(bot, update)
	case "help":
		SendMessage(bot, update.Message.Chat.ID, "Use these commands:\n /auth - to authorize\n /account - to manage your account\n /lists - to manage your todo-lists and tasks")
	default:
		SendMessage(bot, update.Message.Chat.ID, "Unknown command. Enter /help to see all available commands.")
	}
}

func HandleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := GetChatID(update)
	session := GetSession(chatID)

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
		SendMessage(bot, chatID, "Enter new name:")
		session.Status = "edit_name"
	case "edit_name_no":
		session.AccountData = append(session.AccountData, session.UserInfo.Name)
		SendYesNoQuestion(bot, chatID, "Change your username:", "edit_username_yes", "edit_username_no")
	case "edit_username_yes":
		SendMessage(bot, chatID, "Enter new username:")
		session.Status = "edit_username"
	case "edit_username_no":
		session.AccountData = append(session.AccountData, session.UserInfo.Username)
		SendYesNoQuestion(bot, chatID, "Change your password:", "edit_password_yes", "edit_password_no")
	case "edit_password_yes":
		SendMessage(bot, chatID, "Enter new password:")
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
			SendMessage(bot, chatID, "Unable to delete. Try again.")
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to delete: %v", string(response))
			SendMessage(bot, chatID, "Unknown error. Try again.")
			return
		}

		SendMessage(bot, chatID, "Account was deleted successfully.")
		session.UserToken = ""
		session.UserInfo = todo.User{}
		delete(userSessions, chatID)
	case "delete_acc_no":
		SendMessage(bot, chatID, "Operation was canceled.")
	case "log_out":
		LogOut(bot, update)
	case "manage_lists":
		ManageLists(bot, update, session)
	case "edit_list_tasks":
		PromptListSelection(bot, chatID, session, "edit_tasks")
	case "manage_list_info":
		PromptListSelection(bot, chatID, session, "manage_info")
	case "delete_list":
		PromptListSelection(bot, chatID, session, "delete_list")
	case "edit_task_title_yes":
		SendMessage(bot, chatID, "Enter new title:")
		session.Status = "edit_task_title"
	case "edit_task_title_no":
		SendYesNoQuestion(bot, chatID, "Change description:", "edit_task_description_yes", "edit_task_description_no")
	case "edit_task_description_yes":
		SendMessage(bot, chatID, "Enter new description:")
		session.Status = "edit_task_description"
	case "edit_task_description_no":
		SendYesNoQuestion(bot, chatID, "Mark task as completed/uncompleted?", "edit_task_status_yes", "edit_task_status_no")
	case "edit_task_status_yes":
		session.TaskData.Done = !session.TaskData.Done
		session.Status = "edit_task_status"
		HandleUpdateTaskMessage(bot, update, session)
	case "edit_task_status_no":
		session.Status = "edit_task_status"
		HandleUpdateTaskMessage(bot, update, session)
	case "add_list":
		AddList(bot, update)
	default:
		if callbackData := strings.Split(update.CallbackQuery.Data, "|"); len(callbackData) == 2 {
			action := callbackData[0]
			id := callbackData[1]

			switch action {
			case "edit_tasks":
				EditListTasks(bot, chatID, id, session)
			//case "manage_info":
			//	ManageListInfo(bot, chatID, id)
			//case "delete_list":
			//	DeleteList(bot, chatID, id)
			case "add_task", "update_task", "delete_task":
				session.CurrentListID = id
				HandleTaskAction(bot, chatID, session, action, id)
			case "processing_updating_task":
				UpdateTask(bot, chatID, id, session)
			case "processing_deleting_task":
				DeleteTask(bot, chatID, id, session)
			default:
				SendMessage(bot, chatID, "Unknown action.")
			}
		}
	}
}
