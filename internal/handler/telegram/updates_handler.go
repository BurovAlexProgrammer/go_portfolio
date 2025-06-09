package telegram

import (
	"GoPortfolio/internal/service"
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"log/slog"
	"strings"
	"sync"
)

type TgUserState int8

const (
	Undefined       TgUserState = 0
	Default         TgUserState = 1
	WaitAddTask     TgUserState = 2
	WaitDoneTask    TgUserState = 3
	AskNewTaskName  TgUserState = 4
	AskDoneTaskName TgUserState = 5
)

type UpdatesHandler struct {
	bot         *tgbotapi.BotAPI
	authService *service.AuthService
	userStates  map[string]TgUserState
}

func NewUpdatesHandler(bot *tgbotapi.BotAPI, authService *service.AuthService) *UpdatesHandler {
	return &UpdatesHandler{
		bot:         bot,
		authService: authService,
		userStates:  make(map[string]TgUserState, 16),
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
				msg := update.Message
				if isCommand(msg, "/start") {
					h.authService.RegisterByTelegramIfNecessary(context.Background(), msg)
					h.setUserState(msg.From.UserName, Default)
				}

				h.stateLogic(msg)
				//sendText(bot, msg.Chat.ID, "Неверная команда")
			}
		}

		wg.Done()
	}()
}

func (h *UpdatesHandler) stateLogic(msg *tgbotapi.Message) {
	userName := msg.From.UserName
	chatId := msg.Chat.ID

	switch msg.Command() {
	case "newTask":
		h.setUserState(userName, AskNewTaskName)
	case "doneTask":
		h.setUserState(userName, AskDoneTaskName)
	}

	state := h.userStates[userName]

	switch state {
	case Undefined:
		h.sendText(chatId, "Что-то пошло не так")
		h.setUserState(userName, Default)
		h.stateLogic(msg)
	case Default:
		newMsg := tgbotapi.NewMessage(chatId, "Что будем делать?")
		h.addDefaultKeyboard(&newMsg)
		h.sendMessage(&newMsg)
	case WaitAddTask:
		slog.Warn("Задача для добавления: " + msg.Text)
		h.setUserState(userName, Default)
		h.stateLogic(msg)
	case WaitDoneTask:
		slog.Warn("Задача для завершения: " + msg.Text)
		h.setUserState(userName, Default)
		h.stateLogic(msg)
	case AskDoneTaskName:
		newMsg := tgbotapi.NewMessage(chatId, "Выберите задачу для завершения:")
		h.addTasksKeyboard(&newMsg)
		h.sendMessage(&newMsg)
		h.setUserState(userName, WaitDoneTask)
	case AskNewTaskName:
		newMsg := tgbotapi.NewMessage(chatId, "Введите название задачи")
		h.sendMessage(&newMsg)
		h.setUserState(userName, WaitAddTask)
	}
}

func (h *UpdatesHandler) setUserState(userName string, state TgUserState) {
	h.userStates[userName] = state
}

func (h *UpdatesHandler) addDefaultKeyboard(msg *tgbotapi.MessageConfig) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить задачу", "/newTask"),
			tgbotapi.NewInlineKeyboardButtonData("Завершить задачу", "/doneTask"),
		),
	)

	msg.ReplyMarkup = kb
}

func (h *UpdatesHandler) addTasksKeyboard(msg *tgbotapi.MessageConfig) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Задача 1", "Check check ch"),
			tgbotapi.NewInlineKeyboardButtonData("Задача2", "/FUCKYOU 123"),
		),
	)

	msg.ReplyMarkup = kb
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
