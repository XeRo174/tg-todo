package utils

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"strings"
	"tg-todo/internal/types"
	"unicode"
)

// FirstTitleLetter - преобразует первый символ строки к верхнему регистру, а последующие к нижнему
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

// ThemeStroke - формирует из массива тем строку названий тем
func ThemeStroke(themes []types.ThemeModel) string {
	var strokes []string
	for _, theme := range themes {
		strokes = append(strokes, fmt.Sprintf("%s", theme.Name))
	}
	return strings.Join(strokes, ", ")
}

// Contains - проверяет есть ли тема с указанным id в массиве тем
func Contains(themes []types.ThemeModel, themeId uint) bool {
	for _, theme := range themes {
		if theme.ID == themeId {
			return true
		}
	}
	return false
}

// TaskInlineKeyboard - формирует клавиатуру работы с задачей
func TaskInlineKeyboard(nextField string, allowSkip bool) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	if allowSkip {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: "Завершить", CallbackData: types.CallbackTaskComplete},
			{Text: "Пропуск поля", CallbackData: fmt.Sprintf("%s%s", types.CallbackTaskFieldSkip, nextField)},
			{Text: "Прекратить", CallbackData: types.CallbackTaskStop},
		})
	} else {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: "Завершить", CallbackData: types.CallbackTaskComplete},
			{Text: "Прекратить", CallbackData: types.CallbackTaskStop},
		})
	}
	return buttons
}

// PriorityButtons - формирует клавиатуру возможных приоритетов задачи
func PriorityButtons() [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, priority := range types.AllPriorities() {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: priority.String(), CallbackData: fmt.Sprintf("%s%d", types.CallbackTaskPrioritySet, priority)},
		})
	}
	return buttons
}

// ChooseThemeInlineKeyboard - формирует клавиатуру выбора тем
func ChooseThemeInlineKeyboard(themesByPage []types.ThemeModel, totalPages, currentPage int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, theme := range themesByPage {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: theme.Name, CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeChoose, theme.ID)},
		})
	}
	if totalPages > 1 {
		buttons = append(buttons, CreateArrowButtons(currentPage, totalPages)...)
	}
	return buttons
}

// ChooseTaskInlineKeyboard - формирует клавиатуру выбора задач
func ChooseTaskInlineKeyboard(tasksByPage []types.TaskModel, totalPages, currentPage int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, task := range tasksByPage {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: task.Name, CallbackData: fmt.Sprintf("%s%d", types.CallbackTaskChoose, task.ID)},
		})
	}
	if totalPages > 1 {
		buttons = append(buttons, CreateArrowButtons(currentPage, totalPages)...)
	}
	return buttons
}

// ChooseThemeForTaskInlineKeyboard - формирует клавиатуру выбора тем задачи
func ChooseThemeForTaskInlineKeyboard(task types.TaskModel, themesByPage []types.ThemeModel, totalPages, currentPage int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, theme := range themesByPage {
		if Contains(task.Themes, theme.ID) {
			buttonText := fmt.Sprintf("[x] %s", theme.Name)
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: buttonText, CallbackData: fmt.Sprintf("%s%d;%s%d", types.CallbackTaskUnsetTheme, theme.ID, types.CallbackCurrentPage, currentPage)},
			})
		} else {
			buttonText := fmt.Sprintf("[  ] %s", theme.Name)
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: buttonText, CallbackData: fmt.Sprintf("%s%d;%s%d", types.CallbackTaskSetTheme, theme.ID, types.CallbackCurrentPage, currentPage)},
			})
		}
	}
	if totalPages > 1 {
		buttons = append(buttons, CreateArrowButtons(currentPage, totalPages)...)
	}
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Завершить выбор", CallbackData: types.CallbackTaskSetThemeDone}},
	)
	return buttons
}

// CreateArrowButtons - формирует клавиатуру из стрелок переключения страниц
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

// CreateCalendarButtons - создание клавиатуры для сроков задач
func CreateCalendarButtons() [][]gotgbot.InlineKeyboardButton {
	return nil
}
