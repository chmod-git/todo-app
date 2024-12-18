package tg_bot

import (
	"encoding/json"
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"strconv"
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

	var messageText string
	for i, list := range lists {
		messageText += fmt.Sprintf("%d. %s - %s\n", i+1, list.Title, list.Description)
	}

	SendMessage(bot, chatID, messageText)

	editTasksButton := tgbotapi.NewInlineKeyboardButtonData("Edit list tasks", "edit_list_tasks")
	manageInfoButton := tgbotapi.NewInlineKeyboardButtonData("Manage list info", "manage_list_info")
	deleteListButton := tgbotapi.NewInlineKeyboardButtonData("Delete list", "delete_list")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(editTasksButton),
		tgbotapi.NewInlineKeyboardRow(manageInfoButton),
		tgbotapi.NewInlineKeyboardRow(deleteListButton),
	)

	msg := tgbotapi.NewMessage(chatID, "What would you like to do with your lists?")
	msg.ReplyMarkup = keyboard
	_, err = bot.Send(msg)
	if err != nil {
		logrus.Errorf("Failed to send keyboard: %v", err)
	}
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

func EditListTasks(bot *tgbotapi.BotAPI, chatID int64, listId string, session *UserSession) {
	httpClient := NewHTTPClient("http://localhost:8000")

	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET(fmt.Sprintf("/api/lists/%s/items", listId), headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to retrieve tasks: %v", err)
		SendMessage(bot, chatID, "Failed to retrieve tasks. Please try again.")
		return
	}

	var getAllItemsResponse handler.GetAllItemsResponse
	err = json.Unmarshal(response, &getAllItemsResponse)
	if err != nil {
		logrus.Errorf("Failed to parse tasks response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing tasks.")
		return
	}
	items := getAllItemsResponse.Data

	response, statusCode, err = httpClient.GET(fmt.Sprintf("/api/lists/%s", listId), headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to retrieve lists: %v", err)
		SendMessage(bot, chatID, "Failed to retrieve lists. Please try again.")
		return
	}

	var list todo.TodoList
	err = json.Unmarshal(response, &list)
	if err != nil {
		logrus.Errorf("Failed to parse lists response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing your lists.")
		return
	}

	var messageText string
	messageText += fmt.Sprintf("%s:\n\n", list.Title)
	for i, item := range items {
		statusEmoji := "❌"
		if item.Done {
			statusEmoji = "✅"
		}
		messageText += fmt.Sprintf("%d. %s - %s %s\n", i+1, item.Title, item.Description, statusEmoji)
	}

	addTaskButton := tgbotapi.NewInlineKeyboardButtonData("Add new task", fmt.Sprintf("add_task|%s", listId))
	updateTaskButton := tgbotapi.NewInlineKeyboardButtonData("Update task info", fmt.Sprintf("update_task|%s", listId))
	deleteTaskButton := tgbotapi.NewInlineKeyboardButtonData("Delete item", fmt.Sprintf("delete_task|%s", listId))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(addTaskButton),
		tgbotapi.NewInlineKeyboardRow(updateTaskButton),
		tgbotapi.NewInlineKeyboardRow(deleteTaskButton),
	)

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = keyboard
	_, err = bot.Send(msg)
	if err != nil {
		logrus.Errorf("Failed to send message: %v", err)
	}
}

func PromptTaskSelection(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action string) {
	if action == "update_task" {
		action = "processing_updating_task"
	} else if action == "delete_task" {
		action = "processing_deleting_task"
	}

	httpClient := NewHTTPClient("http://localhost:8000")

	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET(fmt.Sprintf("/api/lists/%s/items", session.CurrentListID), headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to retrieve tasks: %v", err)
		SendMessage(bot, chatID, "Failed to retrieve tasks. Please try again.")
		return
	}

	var getAllItemsResponse handler.GetAllItemsResponse
	err = json.Unmarshal(response, &getAllItemsResponse)
	if err != nil {
		logrus.Errorf("Failed to parse tasks response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing tasks.")
		return
	}
	items := getAllItemsResponse.Data

	if len(items) == 0 {
		SendMessage(bot, chatID, "This list has no tasks.")
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, item := range items {
		button := tgbotapi.NewInlineKeyboardButtonData(item.Title, fmt.Sprintf("%s|%d", action, item.Id))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, "Select a task:")
	msg.ReplyMarkup = keyboard
	_, err = bot.Send(msg)
}

func HandleTaskAction(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action, taskID string) {
	switch action {
	case "add_task":
		AddTask(bot, chatID, session)
	case "update_task":
		PromptTaskSelection(bot, chatID, session, action)
	case "delete_task":
		PromptTaskSelection(bot, chatID, session, action)
	default:
		SendMessage(bot, chatID, "Unknown task action.")
	}
}

func AddTask(bot *tgbotapi.BotAPI, chatID int64, session *UserSession) {
	session.Status = "add_task_title"
	session.AccountData = []string{}

	SendMessage(bot, chatID, "Enter the title:")
}

func HandleAddTaskMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)
	text := update.Message.Text

	switch len(session.AccountData) {
	case 0:
		session.AccountData = append(session.AccountData, text)
		SendMessage(bot, chatID, "Enter the description:")
		session.Status = "add_task_description"
	case 1:
		session.AccountData = append(session.AccountData, text)
		listID, err := strconv.Atoi(session.CurrentListID)
		if err != nil {
			logrus.Errorf("Failed to parse list ID: %v", err)
			SendMessage(bot, chatID, "An error occurred while adding the task. Please try again.")
			return
		}

		task := todo.TodoItem{
			ListId:      listID,
			Title:       session.AccountData[0],
			Description: session.AccountData[1],
			Done:        false,
		}

		httpClient := NewHTTPClient("http://localhost:8000")
		headers := map[string]string{
			"Authorization": "Bearer " + session.UserToken,
		}

		_, statusCode, err := httpClient.POST(fmt.Sprintf("/api/lists/%s/items", session.CurrentListID), headers, task)
		if err != nil || statusCode != 200 {
			logrus.Errorf("Failed to add task: %v", err)
			SendMessage(bot, chatID, "An error occurred while adding the task. Please try again.")
			return
		}

		SendMessage(bot, chatID, "Task added successfully!")
		session.Status = ""
		session.AccountData = nil
		session.CurrentListID = ""
	default:
		SendMessage(bot, chatID, "Error. Please restart by clicking on Add Task.")
		session.Status = ""
		session.AccountData = nil
		session.CurrentListID = ""
	}
}

func UpdateTask(bot *tgbotapi.BotAPI, chatID int64, taskID string, session *UserSession) {
	session.Status = "update_task_title"

	httpClient := NewHTTPClient("http://localhost:8000")
	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET(fmt.Sprintf("/api/lists/%s/items/%s", session.CurrentListID, taskID), headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to retrieve task: %v", err)
		SendMessage(bot, chatID, "Failed to retrieve the task. Please try again.")
		return
	}

	err = json.Unmarshal(response, &session.TaskData)
	if err != nil {
		logrus.Errorf("Failed to parse task response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing the task.")
	}

	SendYesNoQuestion(bot, chatID, "Change title:", "edit_task_title_yes", "edit_task_title_no")
}

func HandleUpdateTaskMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)

	switch session.Status {
	case "edit_task_title":
		session.TaskData.Title = update.Message.Text
		session.Status = "edit_task_description"
		SendYesNoQuestion(bot, chatID, "Change description:", "edit_task_description_yes", "edit_task_description_no")
	case "edit_task_description":
		session.TaskData.Description = update.Message.Text
		session.Status = "edit_task_status"
		SendYesNoQuestion(bot, chatID, "Mark task as completed?", "edit_task_status_yes", "edit_task_status_no")
	case "edit_task_status":
		httpClient := NewHTTPClient("http://localhost:8000")
		headers := map[string]string{
			"Authorization": "Bearer " + session.UserToken,
		}

		input := todo.UpdateItemInput{
			Title:       &session.TaskData.Title,
			Description: &session.TaskData.Description,
			Done:        &session.TaskData.Done,
		}

		response, statusCode, err := httpClient.PUT(fmt.Sprintf("/api/lists/%s/items/%d", session.CurrentListID, session.TaskData.Id), headers, input)
		if err != nil {
			logrus.Errorf("Failed to update task: %v", err)
			SendMessage(bot, chatID, "Failed to update task. Please try again.")
			return
		} else if statusCode != 200 {
			logrus.Errorf("Failed to update task: %v", string(response))
			SendMessage(bot, chatID, "An error occurred while updating the task.")
			return
		}

		SendMessage(bot, chatID, "Task successfully updated!")
		session.TaskData = todo.TodoItem{}
		session.Status = ""
	default:
		SendMessage(bot, chatID, "Unknown operation.")
	}
}

func DeleteTask(bot *tgbotapi.BotAPI, chatID int64, taskID string, session *UserSession) {
	httpClient := NewHTTPClient("http://localhost:8000")

	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	_, statusCode, err := httpClient.DELETE(fmt.Sprintf("/api/lists/%s/items/%s", session.CurrentListID, taskID), headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to delete task: %v", err)
		SendMessage(bot, chatID, "Failed to delete the task. Please try again.")
		return
	}

	SendMessage(bot, chatID, "Task successfully deleted.")
}

func AddList(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// TODO
}
