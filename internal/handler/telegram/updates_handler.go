package telegram

import (
	"GoPortfolio/internal/domain"
	"GoPortfolio/internal/domain/tg/tgCommands"
	"GoPortfolio/internal/domain/tg/tgStates"
	"GoPortfolio/internal/service"
	"GoPortfolio/internal/utility/telegramUtility"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"sync"
)

type UpdatesHandler struct {
	bot            *tgbotapi.BotAPI
	authService    *service.AuthService
	taskService    *service.TaskService
	userStates     map[string]tgStates.TgUserState
	userStateMutex sync.RWMutex
}

func NewUpdatesHandler(bot *tgbotapi.BotAPI, authService *service.AuthService, taskService *service.TaskService) *UpdatesHandler {
	return &UpdatesHandler{
		bot:         bot,
		authService: authService,
		userStates:  make(map[string]tgStates.TgUserState, 16),
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
				if isCommand(msg, tgCommands.Start) {
					h.authService.RegisterByTelegramIfNecessary(ctx, msg.From.UserName, msg.From.FirstName)
					h.setUserState(msg.From.UserName, tgStates.Default)
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

func (h *UpdatesHandler) getUserState(username string) tgStates.TgUserState {
	h.userStateMutex.RLock()
	defer h.userStateMutex.RUnlock()
	return h.userStates[username]
}

func (h *UpdatesHandler) buttonsLogic(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	userName := callback.From.UserName
	userState := h.getUserState(userName)
	chatId := callback.Message.Chat.ID
	data := callback.Data

	if userState == tgStates.WaitDoneTask && data != tgCommands.Done {
		h.stateLogic(ctx, userName, chatId, data)
		return
	}

	switch data {
	case tgCommands.NewTask:
		h.setUserState(userName, tgStates.AskNewTaskName)
	case tgCommands.DoneTask:
		h.setUserState(userName, tgStates.AskDoneTaskName)
	case tgCommands.CleanDoneTasks:
		h.setUserState(userName, tgStates.CleanDoneTasks)
	case tgCommands.CleanAllTasks:
		h.setUserState(userName, tgStates.CleanAllTasks)
	case tgCommands.Done:
		h.setUserState(userName, tgStates.Default)
	default:
		h.setUserState(userName, tgStates.Undefined)
	}

	h.stateLogic(ctx, userName, chatId, data)
}

func (h *UpdatesHandler) stateLogic(ctx context.Context, tgUserName string, chatId int64, data string) {
	state := h.getUserState(tgUserName)
	currUser, _ := h.authService.GetExistUser(ctx, tgUserName)
	if currUser == nil {
		err := h.authService.RegisterByTelegramIfNecessary(ctx, tgUserName, "")
		if err != nil {
			h.showError(chatId, err, "Не удалось зарегистрироваться")
		}
		currUser, _ = h.authService.GetExistUser(ctx, tgUserName)
	}

	switch state {
	case tgStates.Undefined:
		h.sendText(chatId, "Что-то пошло не так")
		h.setUserState(tgUserName, tgStates.Default)
		h.stateLogic(ctx, tgUserName, chatId, "")
	case tgStates.Default:
		h.showTaskList(ctx, chatId, currUser)
		newMsg := tgbotapi.NewMessage(chatId, "Что будем делать?")
		telegramUtility.AddDefaultKeyboard(&newMsg)
		h.sendMessage(&newMsg)
	case tgStates.WaitAddTask:
		slog.Info("Задача для добавления: " + data)
		_, err := h.taskService.Create(ctx, data, currUser.Id)
		h.ifErrorShow(chatId, err, fmt.Sprintf("Не удалось создать задачу. taskName:%s, userId:%d", data, currUser.Id))
		h.showTaskList(ctx, chatId, currUser)
		newMsg := tgbotapi.NewMessage(chatId, "Добавим ещё?")
		telegramUtility.AddDoneKeyboard(&newMsg)
		h.sendMessage(&newMsg)
	case tgStates.WaitDoneTask:
		slog.Warn("Задача для завершения: " + data)
		err := h.taskService.DoneByName(ctx, data, currUser.Id)
		h.ifErrorShow(chatId, err, fmt.Sprintf("Не удалось завершить задачу. taskName:%s, userId:%d", data, currUser.Id))
		h.setUserState(tgUserName, tgStates.AskDoneTaskName)
		h.stateLogic(ctx, tgUserName, chatId, "")
	case tgStates.AskDoneTaskName:
		tasks, err := h.taskService.ListByUser(ctx, currUser.Id)
		h.ifErrorShow(chatId, err, "Херня какая-то вышла. Давай по новой (:")
		newMsg := tgbotapi.NewMessage(chatId, "Выберите задачу для завершения:")
		telegramUtility.AddTasksKeyboard(&newMsg, tasks)
		h.sendMessage(&newMsg)
		h.setUserState(tgUserName, tgStates.WaitDoneTask)
	case tgStates.AskNewTaskName:
		newMsg := tgbotapi.NewMessage(chatId, "Введите название задачи")
		h.sendMessage(&newMsg)
		h.setUserState(tgUserName, tgStates.WaitAddTask)
	case tgStates.CleanDoneTasks:
		err := h.taskService.CleanDoneTasksByUserId(ctx, currUser.Id)
		h.ifErrorShow(chatId, err, "Не удалось очистить выполненные задачи")
		h.setUserState(tgUserName, tgStates.Default)
		h.stateLogic(ctx, tgUserName, chatId, "")
	case tgStates.CleanAllTasks:
		err := h.taskService.CleanTasksByUserId(ctx, currUser.Id)
		h.ifErrorShow(chatId, err, "Не удалось удалить все задачи")
		h.setUserState(tgUserName, tgStates.Default)
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

func (h *UpdatesHandler) setUserState(userName string, state tgStates.TgUserState) {
	h.userStateMutex.Lock()
	h.userStates[userName] = state
	h.userStateMutex.Unlock()
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
		names[i] = strconv.Itoa(i+1) + " - " + task.Name + doneMark
	}

	str := strings.Join(names, "\n")
	return str, nil
}

func (h *UpdatesHandler) showTaskList(ctx context.Context, chatId int64, currUser *domain.User) {
	tasksMessage, err := h.tasksListByUser(ctx, currUser.Id)
	tasksMessage = "Список задач: \n" + tasksMessage
	h.ifErrorShow(chatId, err, fmt.Sprintf("Не удалось получить список задач, userId:%d", currUser.Id))
	tasksMsg := tgbotapi.NewMessage(chatId, tasksMessage)
	h.sendMessage(&tasksMsg)
}
