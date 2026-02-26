package utils

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"tg-todo/internal/types"
)

const (
	ConversationThemeCreateName    = "theme_name"
	ConversationTaskCreateName     = "task_name"
	ConversationTaskCreatePriority = "task_priority"
	ConversationTaskCreateTheme    = "task_theme"
	ConversationTaskCreateDeadline = "task_deadline"

	TimeLayout = "2006-01-02 15:04:05"
)

func PriorityButtons() [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, priority := range types.Priorities {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: priority.Name, CallbackData: fmt.Sprintf("task_priority:%d", priority.Value)},
		})
	}
	return buttons
}
