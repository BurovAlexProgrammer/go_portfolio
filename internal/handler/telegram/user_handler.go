package telegram

import (
	"GoPortfolio/internal/usecase"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"log/slog"
	"sync"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: uc,
	}
}

func (h UserHandler) StartUpdates(bot *tgbotapi.BotAPI, wg *sync.WaitGroup) {
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
				slog.Info("[%s] %s", update.Message.From.UserName, update.Message.Text)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID

				if _, err := bot.Send(msg); err != nil {
					log.Printf("Failed to send message: %v", err)
				}

				if update.Message.Text == "/exit" {
					log.Println("Bot stopped")
					break
				}
			}
		}

		wg.Done()
	}()
}

//func (h *UserHandler) CreateUser(ctx *gin.Context) {
//	const op = "http.Handler.CreateUser"
//	slog.Info("Request createUser", "method", ctx.Request.Method, "url", ctx.Request.URL.String())
//
//	var user domain.User
//	if err := ctx.ShouldBindJSON(&user); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
//	if err := h.userUsecase.CreateUser(ctx.Request.Context(), &user); err != nil {
//		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
//		return
//	}
//
//	ctx.JSON(http.StatusCreated, user)
//}
