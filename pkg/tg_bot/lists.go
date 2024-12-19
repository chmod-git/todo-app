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

type ListService struct {
	auth  *AuthorizationService
	redis *repository.RedisRepository
}

func NewListService(auth *AuthorizationService, redis *repository.RedisRepository) *ListService {
	return &ListService{
		auth:  auth,
		redis: redis,
	}
}

func (l *ListService) Lists(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	session := GetSession(chatID)

	if session.UserToken == "" {
		l.auth.Auth(bot, update)
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

func (l *ListService) ManageLists(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
	chatID := GetChatID(update)

	lists, err := l.redis.GetListsData(strconv.FormatInt(chatID, 10))
	if err == nil && len(lists) > 0 {
		sendListsToUser(bot, chatID, lists)
		return
	}

	httpClient := NewHTTPClient("http://localhost:8000")

	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET("/api/lists", headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to manage lists: %v", err)
		SendMessage(bot, chatID, "An error occurred while retrieving your lists. Please try again.")
		return
	}

	var getAllListsResponse handler.GetAllListsResponse
	err = json.Unmarshal(response, &getAllListsResponse)
	if err != nil {
		logrus.Errorf("Failed to unmarshal response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing your lists.")
		return
	}

	lists = getAllListsResponse.Data

	if len(lists) == 0 {
		SendMessage(bot, chatID, "You don't have any lists yet.")
		return
	}

	err = l.redis.SaveListsData(strconv.FormatInt(chatID, 10), lists, 900)
	if err != nil {
		logrus.Errorf("Failed to save lists to Redis: %v", err)
	}

	sendListsToUser(bot, chatID, lists)
}

func sendListsToUser(bot *tgbotapi.BotAPI, chatID int64, lists []todo.TodoList) {
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

func (l *ListService) PromptListSelection(bot *tgbotapi.BotAPI, chatID int64, session *UserSession, action string) {
	lists, err := l.redis.GetListsData(strconv.FormatInt(chatID, 10))
	if err == nil && len(lists) > 0 {
		sendListsChoiceToUser(bot, chatID, lists, action)
		return
	}

	httpClient := NewHTTPClient("http://localhost:8000")

	headers := map[string]string{
		"Authorization": "Bearer " + session.UserToken,
	}

	response, statusCode, err := httpClient.GET("/api/lists", headers)
	if err != nil || statusCode != 200 {
		logrus.Errorf("Failed to manage lists: %v", err)
		SendMessage(bot, chatID, "An error occurred while retrieving your lists. Please try again.")
		return
	}

	var getAllListsResponse handler.GetAllListsResponse
	err = json.Unmarshal(response, &getAllListsResponse)
	if err != nil {
		logrus.Errorf("Failed to unmarshal response: %v", err)
		SendMessage(bot, chatID, "An error occurred while processing your lists.")
		return
	}

	lists = getAllListsResponse.Data

	if len(lists) == 0 {
		SendMessage(bot, chatID, "You don't have any lists yet.")
		return
	}

	sendListsChoiceToUser(bot, chatID, lists, action)
}

func sendListsChoiceToUser(bot *tgbotapi.BotAPI, chatID int64, lists []todo.TodoList, action string) {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, list := range lists {
		button := tgbotapi.NewInlineKeyboardButtonData(list.Title, fmt.Sprintf("%s|%d", action, list.Id))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, "Select a list:")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func (l *ListService) ManageListInfo(bot *tgbotapi.BotAPI, chatID int64, listID string, session *UserSession) {
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

func (l *ListService) HandleManageListInfoMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
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

		err = l.redis.UpdateListData(fmt.Sprintf("%d", chatID), session.ListData, 900)
		if err != nil {
			logrus.Errorf("Failed to update list in Redis: %v", err)
			SendMessage(bot, chatID, "List updated on server, but failed to sync with Redis.")
			return
		}

		SendMessage(bot, chatID, "List successfully updated!")
		session.TaskData = todo.TodoItem{}
		session.Status = ""
	default:
		SendMessage(bot, chatID, "Unknown operation.")
	}
}

func (l *ListService) DeleteList(bot *tgbotapi.BotAPI, chatID int64, listID string, session *UserSession) {
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

	redisKey := strconv.FormatInt(chatID, 10)
	id, _ := strconv.Atoi(listID)
	err = l.redis.DeleteListData(redisKey, todo.TodoList{Id: id}, 900)
	if err != nil {
		logrus.Warnf("Failed to update Redis cache after deleting list: %v", err)
	}

	SendMessage(bot, chatID, "List successfully deleted.")
}

func (l *ListService) AddList(bot *tgbotapi.BotAPI, chatID int64, session *UserSession) {
	session.Status = "add_list_title"
	session.AccountData = []string{}

	SendMessage(bot, chatID, "Enter the title:")
}

func (l *ListService) HandleAddListMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, session *UserSession) {
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

		response, statusCode, err := httpClient.POST("/api/lists", headers, list)
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

		listID, ok := responseData["id"].(float64)
		if !ok {
			logrus.Errorf("Invalid response format: missing 'id'")
			SendMessage(bot, chatID, "Invalid server response. Please try again.")
			return
		}

		list.Id = int(listID)

		err = l.redis.AddListData(fmt.Sprintf("%d", chatID), list, 900)
		if err != nil {
			logrus.Errorf("Failed to sync with Redis: %v", err)
			SendMessage(bot, chatID, "An error occurred while syncing with Redis. Please try again.")
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
