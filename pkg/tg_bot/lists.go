package tg_bot

import (
	"encoding/json"
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func Lists(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	session := GetSession(chatID)

	if session.UserToken == "" {
		Auth(bot, update)
		return
	}

	signUpButton := tgbotapi.NewInlineKeyboardButtonData("Manage existing lists", "manage_lists")
	signInButton := tgbotapi.NewInlineKeyboardButtonData("Add new list", "add_list")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(signUpButton),
		tgbotapi.NewInlineKeyboardRow(signInButton),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Manage your lists:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func ManageLists(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)
	httpClient := NewHTTPClient("http://localhost:8000")

	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET("/api/lists", headers)
	if err != nil {
		logrus.Errorf("Failed to manage lists: %v", err)
		SendMessage(bot, chatID, "Failed to retrieve your lists. Please try again.")
		return
	} else if statusCode != 200 {
		logrus.Errorf("Failed to manage lists: %v", string(response))
		SendMessage(bot, chatID, "An error occurred. Please try again.")
		return
	}

	var getAllListsResponse handler.GetAllListsResponse
	err = json.Unmarshal(response, &getAllListsResponse)
	if err != nil {
		logrus.Errorf("Failed to unmarshal response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing your lists.")
		return
	}
	lists := getAllListsResponse.Data

	if len(lists) == 0 {
		SendMessage(bot, chatID, "You don't have any lists yet.")
		return
	}

	var messageText = "Your lists:\n\n"
	for i, list := range lists {
		messageText += fmt.Sprintf("%d. %s - %s\n", i+1, list.Title, list.Description)
	}

	editTasksButton := tgbotapi.NewInlineKeyboardButtonData("Edit list tasks", "edit_list_tasks")
	manageInfoButton := tgbotapi.NewInlineKeyboardButtonData("Manage list info", "manage_list_info")
	deleteListButton := tgbotapi.NewInlineKeyboardButtonData("Delete list", "delete_list")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(editTasksButton),
		tgbotapi.NewInlineKeyboardRow(manageInfoButton),
		tgbotapi.NewInlineKeyboardRow(deleteListButton),
	)

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func PromptListSelection(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action string) {
	httpClient := NewHTTPClient("http://localhost:8000")

	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET("/api/lists", headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to retrieve lists: %v", err)
		SendMessage(bot, chatID, "Failed to retrieve lists. Please try again.")
		return
	}

	var getAllListsResponse handler.GetAllListsResponse
	err = json.Unmarshal(response, &getAllListsResponse)
	if err != nil {
		logrus.Errorf("Failed to unmarshal response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing your lists.")
		return
	}

	lists := getAllListsResponse.Data
	if len(lists) == 0 {
		SendMessage(bot, chatID, "You don't have any lists yet.")
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, list := range lists {
		button := tgbotapi.NewInlineKeyboardButtonData(list.Title, fmt.Sprintf("%s|%d", action, list.Id))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, "Select a list:")
	msg.ReplyMarkup = keyboard
	_, err = bot.Send(msg)
}

func ManageListInfo(bot *tgbotapi.BotAPI, chatID int64, listID string, session *UserSession) {
	httpClient := NewHTTPClient("http://localhost:8000")
	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET(fmt.Sprintf("/api/lists/%s", listID), headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to retrieve list: %v", err)
		SendMessage(bot, chatID, "Failed to retrieve the list. Please try again.")
		return
	}

	err = json.Unmarshal(response, &session.ListData)
	if err != nil {
		logrus.Errorf("Failed to parse list response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing the list.")
	}

	session.CurrentListID = listID
	SendYesNoQuestion(bot, chatID, "Change title:", "edit_list_title_yes", "edit_list_title_no")
}

func HandleManageListInfoMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)

	switch session.Status {
	case "edit_list_title":
		session.ListData.Title = update.Message.Text
		session.Status = "edit_list_description"
		SendYesNoQuestion(bot, chatID, "Change description:", "edit_list_description_yes", "edit_list_description_no")
	case "edit_list_description":
		session.ListData.Description = update.Message.Text

		httpClient := NewHTTPClient("http://localhost:8000")
		headers := map[string]string{
			"Authorization": "Bearer " + session.UserToken,
		}

		input := todo.UpdateListInput{
			Title:       &session.ListData.Title,
			Description: &session.ListData.Description,
		}

		response, statusCode, err := httpClient.PUT(fmt.Sprintf("/api/lists/%s", session.CurrentListID), headers, input)
		if err != nil {
			logrus.Errorf("Failed to update list: %v", err)
			SendMessage(bot, chatID, "Failed to update list. Please try again.")
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to update list: %v", string(response))
			SendMessage(bot, chatID, "An error occurred while updating the list.")
			return
		}

		SendMessage(bot, chatID, "List successfully updated!")
		session.TaskData = todo.TodoItem{}
		session.Status = ""
	default:
		SendMessage(bot, chatID, "Unknown operation.")
	}
}

func DeleteList(bot *tgbotapi.BotAPI, chatID int64, listID string, session *UserSession) {
	httpClient := NewHTTPClient("http://localhost:8000")

	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	_, statusCode, err := httpClient.DELETE(fmt.Sprintf("/api/lists/%s", listID), headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to delete list: %v", err)
		SendMessage(bot, chatID, "Failed to delete the list. Please try again.")
		return
	}

	session.CurrentListID = listID
	SendMessage(bot, chatID, "List successfully deleted.")
}

func AddList(bot *tgbotapi.BotAPI, chatID int64, session *UserSession) {
	session.Status = "add_list_title"
	session.AccountData = []string{}

	SendMessage(bot, chatID, "Enter the title:")
}

func HandleAddListMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)
	text := update.Message.Text

	switch len(session.AccountData) {
	case 0:
		session.AccountData = append(session.AccountData, text)
		SendMessage(bot, chatID, "Enter the description:")
		session.Status = "add_list_description"
	case 1:
		session.AccountData = append(session.AccountData, text)

		list := todo.TodoList{
			Title:       session.AccountData[0],
			Description: session.AccountData[1],
		}

		httpClient := NewHTTPClient("http://localhost:8000")
		headers := map[string]string{
			"Authorization": "Bearer " + session.UserToken,
		}

		_, statusCode, err := httpClient.POST("/api/lists", headers, list)
		if err != nil || statusCode != 200 {
			logrus.Errorf("Failed to add task: %v", err)
			SendMessage(bot, chatID, "An error occurred while adding the task. Please try again.")
			return
		}

		SendMessage(bot, chatID, "List added successfully!")
		session.Status = ""
		session.AccountData = nil
		session.CurrentListID = ""
	default:
		SendMessage(bot, chatID, "Error. Please restart by clicking on Add List.")
		session.Status = ""
		session.AccountData = nil
		session.CurrentListID = ""
	}
}
