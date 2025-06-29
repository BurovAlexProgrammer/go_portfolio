package telegram

import (
	"GoPortfolio/internal/domain"
	"GoPortfolio/internal/service"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"sync"
)

type TgUserState int8
type TgCommand string

const (
	Undefined       TgUserState = 0
	Default         TgUserState = 1
	WaitAddTask     TgUserState = 2
	WaitDoneTask    TgUserState = 3
	AskNewTaskName  TgUserState = 4
	AskDoneTaskName TgUserState = 5
	CleanDoneTasks  TgUserState = 6
	CleanAllTasks   TgUserState = 7

	startCmd          string = "/start"
	newTaskCmd        string = "/newTask"
	doneTaskCmd       string = "/doneTask"
	cleanDoneTasksCmd string = "/cleanDoneTasks"
	cleanAllTasksCmd  string = "/cleanAllTasks"
)

type UpdatesHandler struct {
	bot         *tgbotapi.BotAPI
	authService *service.AuthService
	taskService *service.TaskService
	userStates  map[string]TgUserState
}

func NewUpdatesHandler(bot *tgbotapi.BotAPI, authService *service.AuthService, taskService *service.TaskService) *UpdatesHandler {
	return &UpdatesHandler{
		bot:         bot,
		authService: authService,
		userStates:  make(map[string]TgUserState, 16),
		taskService: taskService,
	}
}

func (h *UpdatesHandler) StartUpdates(wg *sync.WaitGroup) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates, err := h.bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatalf("Failed to get updates: %v", err)
	}

	wg.Add(1)
	slog.Info("Telegram bot started, waiting for updates...")

	go func() {

		for update := range updates {
			if update.Message != nil {
				ctx := context.Background()
				msg := update.Message
				if isCommand(msg, startCmd) {
					h.authService.RegisterByTelegramIfNecessary(ctx, msg.From.UserName, msg.From.FirstName)
					h.setUserState(msg.From.UserName, Default)
				}

				h.stateLogic(ctx, msg.From.UserName, msg.Chat.ID, msg.Text)
				continue
			}
			if update.CallbackQuery != nil {
				ctx := context.Background()
				h.buttonsLogic(ctx, update.CallbackQuery)
				continue
			}
		}

		wg.Done()
	}()
}

func (h *UpdatesHandler) buttonsLogic(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	userName := callback.From.UserName
	chatId := callback.Message.Chat.ID
	data := callback.Data

	switch data {
	case newTaskCmd:
		h.setUserState(userName, AskNewTaskName)
	case doneTaskCmd:
		h.setUserState(userName, AskDoneTaskName)
	case cleanDoneTasksCmd:
		h.setUserState(userName, CleanDoneTasks)
	case cleanAllTasksCmd:
		h.setUserState(userName, CleanAllTasks)
	default:
		h.setUserState(userName, Undefined)
	}

	h.stateLogic(ctx, userName, chatId, data)
}

func (h *UpdatesHandler) stateLogic(ctx context.Context, tgUserName string, chatId int64, data string) {
	state := h.userStates[tgUserName]
	currUser, _ := h.authService.GetExistUser(ctx, tgUserName)
	if currUser == nil {
		err := h.authService.RegisterByTelegramIfNecessary(ctx, tgUserName, "")
		if err != nil {
			h.showError(chatId, err, "Не удалось зарегистрироваться")
		}
		currUser, _ = h.authService.GetExistUser(ctx, tgUserName)
	}

	switch state {
	case Undefined:
		h.sendText(chatId, "Что-то пошло не так")
		h.setUserState(tgUserName, Default)
		h.stateLogic(ctx, tgUserName, chatId, "")
	case Default:
		tasksMessage, err := h.tasksListByUser(ctx, currUser.Id)
		h.ifErrorShow(chatId, err, fmt.Sprintf("Не удалось получить список задач, userId:%d", currUser.Id))
		tasksMsg := tgbotapi.NewMessage(chatId, tasksMessage)
		h.sendMessage(&tasksMsg)
		newMsg := tgbotapi.NewMessage(chatId, "Что будем делать?")
		h.addDefaultKeyboard(&newMsg)
		h.sendMessage(&newMsg)
	case WaitAddTask:
		slog.Info("Задача для добавления: " + data)
		_, err := h.taskService.Create(ctx, data, currUser.Id)
		h.ifErrorShow(chatId, err, fmt.Sprintf("Не удалось создать задачу. taskName:%s, userId:%d", data, currUser.Id))
		h.setUserState(tgUserName, Default)
		h.stateLogic(ctx, tgUserName, chatId, "")
	case WaitDoneTask:
		slog.Warn("Задача для завершения: " + data)
		err := h.taskService.DoneByName(ctx, data, currUser.Id)
		h.ifErrorShow(chatId, err, fmt.Sprintf("Не удалось завершить задачу. taskName:%s, userId:%d", data, currUser.Id))
		h.setUserState(tgUserName, Default)
		h.stateLogic(ctx, tgUserName, chatId, "")
	case AskDoneTaskName:
		tasks, err := h.taskService.ListByUser(ctx, currUser.Id)
		h.ifErrorShow(chatId, err, "Херня какая-то вышла. Давай по новой (:")
		newMsg := tgbotapi.NewMessage(chatId, "Выберите задачу для завершения:")
		h.addTasksKeyboard(&newMsg, tasks)
		h.sendMessage(&newMsg)
		h.setUserState(tgUserName, WaitDoneTask)
	case AskNewTaskName:
		newMsg := tgbotapi.NewMessage(chatId, "Введите название задачи")
		h.sendMessage(&newMsg)
		h.setUserState(tgUserName, WaitAddTask)
	case CleanDoneTasks:
		err := h.taskService.CleanDoneTasksByUserId(ctx, currUser.Id)
		h.ifErrorShow(chatId, err, "Не удалось очистить выполненные задачи")
		h.setUserState(tgUserName, Default)
		h.stateLogic(ctx, tgUserName, chatId, "")
	case CleanAllTasks:
		err := h.taskService.CleanTasksByUserId(ctx, currUser.Id)
		h.ifErrorShow(chatId, err, "Не удалось удалить все задачи")
		h.setUserState(tgUserName, Default)
		h.stateLogic(ctx, tgUserName, chatId, "")
	}
}

func (h *UpdatesHandler) ifErrorShow(chatId int64, err error, desc string) {
	if err == nil {
		return
	}

	h.showError(chatId, err, desc)
}

func (h *UpdatesHandler) showError(chatId int64, err error, desc string) {
	if desc == "" {
		desc = "Что-то пошло не так"
	}
	slog.Error(fmt.Sprintf("%v:\n %v", desc, err.Error()))
	newMsg := tgbotapi.NewMessage(chatId, "Введите название задачи")
	h.sendMessage(&newMsg)
}

func (h *UpdatesHandler) setUserState(userName string, state TgUserState) {
	h.userStates[userName] = state
}

func (h *UpdatesHandler) addDefaultKeyboard(msg *tgbotapi.MessageConfig) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить задачу", newTaskCmd),
			tgbotapi.NewInlineKeyboardButtonData("Завершить задачу", doneTaskCmd),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Очистить завершенные", cleanDoneTasksCmd),
			tgbotapi.NewInlineKeyboardButtonData("Удалить всё", cleanAllTasksCmd),
		),
	)

	msg.ReplyMarkup = kb
}

// TODO FIX IT
func (h *UpdatesHandler) addTasksKeyboard(msg *tgbotapi.MessageConfig, tasks []domain.Task) {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0)

	for _, t := range tasks {
		button := tgbotapi.NewInlineKeyboardButtonData(t.Name, t.Name) // кнопка с текстом и callback-данными
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *UpdatesHandler) sendText(chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)

	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (h *UpdatesHandler) sendMessage(msg *tgbotapi.MessageConfig) {
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func isCommand(m *tgbotapi.Message, c string) bool {
	return m.IsCommand() && m.Command() == strings.Trim(c, "/")
}

func (h *UpdatesHandler) tasksListByUser(ctx context.Context, userId int64) (string, error) {
	tasks, err := h.taskService.ListByUser(ctx, userId)
	if err != nil {
		return "nil", err
	}

	names := make([]string, len(tasks))
	for i, task := range tasks {
		doneMark := ""
		if task.IsDone {
			doneMark = " ✅"
		}
		names[i] = strconv.Itoa(i+1) + ") " + task.Name + doneMark
	}

	str := strings.Join(names, "\n")
	return str, nil
}
