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

func GetLastMessageRegister(task types.TaskModel, operation string) (types.MessageRegisterModel, bool) {
	for i := len(task.Messages) - 1; i >= 0; i-- {
		if task.Messages[i].Operation == operation {
			return task.Messages[i], true
		}
	}
	return types.MessageRegisterModel{}, false
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

func CreateArrowButtons(currentPage, totalPages int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	if currentPage == 1 {
		// Если мы на первой странице: < - ведет на последнюю страницу и > - ведет на следующую страницу
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: fmt.Sprintf("< (%d/%d)", totalPages, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangePage, totalPages)},
			{Text: fmt.Sprintf("(%d/%d) >", currentPage+1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangePage, currentPage+1)},
		})
	} else if currentPage == totalPages {
		//Если мы на последней странице: < - ведет на прошлую страницу и > - ведет на первую страницу
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangePage, currentPage-1)},
			{Text: fmt.Sprintf("(%d/%d) >", 1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangePage, 1)},
		})
	} else {
		//Иначе: < - ведет на прошлую страницу и > - ведет на следующую страницу
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangePage, currentPage-1)},
			{Text: fmt.Sprintf("(%d/%d) >", currentPage+1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangePage, currentPage+1)},
		})
	}
	return buttons
}
