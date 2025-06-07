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

type UpdatesHandler struct {
	authService *service.AuthService
}

func NewUpdatesHandler(authService *service.AuthService) *UpdatesHandler {
	return &UpdatesHandler{
		authService: authService,
	}
}

func (h UpdatesHandler) StartUpdates(bot *tgbotapi.BotAPI, wg *sync.WaitGroup) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatalf("Failed to get updates: %v", err)
	}

	wg.Add(1)
	slog.Info("Telegram bot started, waiting for updates...")

	go func() {

		for update := range updates {
			if update.Message != nil {
				message := update.Message
				if isCommand(message, "/start") {
					h.authService.RegisterByTelegramIfNecessary(context.Background(), message)
					continue
				}

				sendMessage(bot, message.Chat.ID, "Неверная команда")
			}
		}

		wg.Done()
	}()
}

func isCommand(m *tgbotapi.Message, c string) bool {
	return m.IsCommand() && m.Command() == strings.Trim(c, "/")
}

func sendMessage(bot *tgbotapi.BotAPI, chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
