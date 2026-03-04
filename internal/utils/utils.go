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

func TaskButtons(nextField string) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Создать", CallbackData: types.CallbackTaskCreateDone},                                        // Создать задачу
		{Text: "Пропустить поле", CallbackData: fmt.Sprintf("%s%s", types.CallbackTaskFieldSkip, nextField)}, //Пропустить поле
		{Text: "Отменить", CallbackData: types.CallbackTaskCreateCancel},                                     //Бросить создание
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

func TaskMessageExist(task types.TaskModel, operation string) (types.MessageRegisterModel, bool) {
	for i := len(task.Messages) - 1; i >= 0; i-- {
		if task.Messages[i].Operation == operation {
			return task.Messages[i], true
		}
	}
	return task.Messages[len(task.Messages)-1], true
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
