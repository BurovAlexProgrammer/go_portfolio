package telegramUtility

import (
	"GoPortfolio/internal/domain"
	"GoPortfolio/internal/domain/tg/tgCommands"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func AddDefaultKeyboard(msg *tgbotapi.MessageConfig) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить задачу", tgCommands.NewTask),
			tgbotapi.NewInlineKeyboardButtonData("✅ Завершить задачу", tgCommands.DoneTask),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Очистить завершенные", tgCommands.CleanDoneTasks),
			tgbotapi.NewInlineKeyboardButtonData("❌️ Удалить всё", tgCommands.CleanAllTasks),
		),
	)

	msg.ReplyMarkup = kb
}

func AddTasksKeyboard(msg *tgbotapi.MessageConfig, tasks []domain.Task) {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0)

	for _, t := range tasks {
		if t.IsDone {
			continue
		}
		button := tgbotapi.NewInlineKeyboardButtonData(t.Name, t.Name) // кнопка с текстом и callback-данными
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	backButton := tgbotapi.NewInlineKeyboardButtonData("↩️ Назад", tgCommands.Done)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(backButton))

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func AddDoneKeyboard(msg *tgbotapi.MessageConfig) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("↩️ Хватит", tgCommands.Done),
		),
	)

	msg.ReplyMarkup = kb
}
