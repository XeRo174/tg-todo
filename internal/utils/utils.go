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
			{Text: priority.String(), CallbackData: fmt.Sprintf("set_task_priority:%d", priority)},
		})
	}
	return buttons
}

func ThemesButton(themes []types.ThemeModel) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, theme := range themes {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: theme.Name, CallbackData: fmt.Sprintf("set_task_theme:%s", theme.Name)},
		})
	}
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "<", CallbackData: fmt.Sprintf("change_theme_page:%d", 1)}, //todo установка страницы
		{Text: ">", CallbackData: fmt.Sprintf("change_theme_page:%d", 1)},
	})
	return buttons
}

func TaskButtons() [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Создать", CallbackData: "task_create_done"},
		{Text: "Отменить", CallbackData: "task_create_cancel"},
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
