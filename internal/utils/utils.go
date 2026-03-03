package utils

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"strings"
	"tg-todo/internal/types"
	"unicode"
)

func FirstTitleLetter(stroke string) string {
	if stroke == "" {
		return ""
	}
	r := []rune(strings.ToLower(stroke))
	first := r[0]
	if unicode.IsLetter(first) {
		r[0] = unicode.ToUpper(first)
	}
	return string(r)
}
func PriorityButtons() [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, priority := range types.AllPriorities() {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: priority.String(), CallbackData: fmt.Sprintf("%s%d", types.CallbackTaskPrioritySet, priority)},
		})
	}
	return buttons
}

func TaskButtons() [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Создать", CallbackData: types.CallbackTaskCreateDone},
		{Text: "Отменить", CallbackData: types.CallbackTaskCreateCancel},
	})
	return buttons
}

func ThemeStroke(themes []types.ThemeModel) string {
	var strokes []string
	for _, theme := range themes {
		strokes = append(strokes, fmt.Sprintf("%s", theme.Name))
	}
	return strings.Join(strokes, ", ")
}

func TaskMessageExist(task types.TaskModel) (int64, bool) {
	if len(task.Messages) == 0 {
		return 0, false
	}
	return task.Messages[len(task.Messages)-1].BotMessageId, true
}

// Contains Вспомогательная функция для проверки наличия элемента в слайсе
func Contains(slice []types.ThemeModel, item string) bool {
	for _, s := range slice {
		if s.Name == item {
			return true
		}
	}
	return false
}
