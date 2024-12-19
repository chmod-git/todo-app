package tg_bot

import (
	"encoding/json"
	"fmt"
	"github.com/chmod-git/todo-app"
	"github.com/chmod-git/todo-app/pkg/handler"
	"github.com/chmod-git/todo-app/pkg/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"strconv"
)

type TaskService struct {
	auth  *AuthorizationService
	redis *repository.RedisRepository
}

func NewTaskService(auth *AuthorizationService, redis *repository.RedisRepository) *TaskService {
	return &TaskService{
		auth:  auth,
		redis: redis,
	}
}

func (t *TaskService) EditListTasks(bot *tgbotapi.BotAPI, chatID int64, listId string, session *UserSession) {
	httpClient := NewHTTPClient("http://localhost:8000")
	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET(fmt.Sprintf("/api/lists/%s", listId), headers)
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

	items, err := t.redis.GetItemsData(strconv.FormatInt(chatID, 10), listId)
	if err == nil && len(items) > 0 {
		sendTasksToUser(bot, chatID, list, listId, items)
		return
	}

	response, statusCode, err = httpClient.GET(fmt.Sprintf("/api/lists/%s/items", listId), headers)
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
	items = getAllItemsResponse.Data

	err = t.redis.SaveItemsData(strconv.FormatInt(chatID, 10), listId, items, 900)
	if err != nil {
		logrus.Errorf("Failed to save tasks to Redis: %v", err)
	}

	sendTasksToUser(bot, chatID, list, listId, items)
}

func sendTasksToUser(bot *tgbotapi.BotAPI, chatID int64, list todo.TodoList, listId string, items []todo.TodoItem) {
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
	bot.Send(msg)
}

func (t *TaskService) PromptTaskSelection(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action string) {
	if action == "update_task" {
		action = "processing_updating_task"
	} else if action == "delete_task" {
		action = "processing_deleting_task"
	}

	items, err := t.redis.GetItemsData(strconv.FormatInt(chatID, 10), session.CurrentListID)
	if err == nil && len(items) > 0 {
		sendTasksChoiceToUser(bot, chatID, items, action)
		return
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
	items = getAllItemsResponse.Data

	if len(items) == 0 {
		SendMessage(bot, chatID, "This list has no tasks.")
		return
	}

	sendTasksChoiceToUser(bot, chatID, items, action)
}

func sendTasksChoiceToUser(bot *tgbotapi.BotAPI, chatID int64, items []todo.TodoItem, action string) {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, item := range items {
		button := tgbotapi.NewInlineKeyboardButtonData(item.Title, fmt.Sprintf("%s|%d", action, item.Id))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, "Select a task:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func (t *TaskService) HandleTaskAction(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action, taskID string) {
	switch action {
	case "add_task":
		t.AddTask(bot, chatID, session)
	case "update_task":
		t.PromptTaskSelection(bot, chatID, session, action)
	case "delete_task":
		t.PromptTaskSelection(bot, chatID, session, action)
	default:
		SendMessage(bot, chatID, "Unknown task action.")
	}
}

func (t *TaskService) AddTask(bot *tgbotapi.BotAPI, chatID int64, session *UserSession) {
	session.Status = "add_task_title"
	session.AccountData = []string{}

	SendMessage(bot, chatID, "Enter the title:")
}

func (t *TaskService) HandleAddTaskMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
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

		response, statusCode, err := httpClient.POST(fmt.Sprintf("/api/lists/%s/items", session.CurrentListID), headers, task)
		if err != nil || statusCode != 200 {
			logrus.Errorf("Failed to add task: %v", err)
			SendMessage(bot, chatID, "An error occurred while adding the task. Please try again.")
			return
		}

		var responseData map[string]interface{}
		err = json.Unmarshal(response, &responseData)
		if err != nil {
			logrus.Errorf("Failed to parse response: %v", err)
			SendMessage(bot, chatID, "An error occurred while processing the response. Please try again.")
			return
		}

		itemID, ok := responseData["id"].(float64)
		if !ok {
			logrus.Errorf("Invalid response format: missing 'id'")
			SendMessage(bot, chatID, "Invalid server response. Please try again.")
			return
		}

		task.Id = int(itemID)

		err = t.redis.AddItemData(fmt.Sprintf("%d", chatID), session.CurrentListID, task, 900)
		if err != nil {
			logrus.Errorf("Failed to sync with Redis: %v", err)
			SendMessage(bot, chatID, "An error occurred while syncing with Redis. Please try again.")
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

func (t *TaskService) UpdateTask(bot *tgbotapi.BotAPI, chatID int64, taskID string, session *UserSession) {
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

func (t *TaskService) HandleUpdateTaskMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
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

		err = t.redis.UpdateItemData(fmt.Sprintf("%d", chatID), session.CurrentListID, session.TaskData, 900)
		if err != nil {
			logrus.Errorf("Failed to update task in Redis: %v", err)
			SendMessage(bot, chatID, "Task updated on server, but failed to sync with Redis.")
			return
		}

		SendMessage(bot, chatID, "Task successfully updated!")
		session.TaskData = todo.TodoItem{}
		session.Status = ""
	default:
		SendMessage(bot, chatID, "Unknown operation.")
	}
}

func (t *TaskService) DeleteTask(bot *tgbotapi.BotAPI, chatID int64, taskID string, session *UserSession) {
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

	redisKey := strconv.FormatInt(chatID, 10)
	id, _ := strconv.Atoi(taskID)
	err = t.redis.DeleteItemData(redisKey, session.CurrentListID, todo.TodoItem{Id: id}, 900)
	if err != nil {
		logrus.Warnf("Failed to update Redis cache after deleting task: %v", err)
	}

	SendMessage(bot, chatID, "Task successfully deleted.")
}
