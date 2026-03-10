package utils

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"strings"
	"tg-todo/internal/types"
	"time"
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
func CreateCalendarButtons(chosenYear, chosenMonthNumber, day int) [][]gotgbot.InlineKeyboardButton {
	chosenMonth := time.Month(chosenMonthNumber)
	var buttons [][]gotgbot.InlineKeyboardButton
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("%d", chosenYear), CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineShowYears, chosenYear)},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: chosenMonth.String(), CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineShowMonths, chosenMonth)},
	})
	weekDayNames := []string{"пн", "вт", "ср", "чт", "пт", "сб", "вс"}
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: weekDayNames[0], CallbackData: types.CallbackEmpty},
		{Text: weekDayNames[1], CallbackData: types.CallbackEmpty},
		{Text: weekDayNames[2], CallbackData: types.CallbackEmpty},
		{Text: weekDayNames[3], CallbackData: types.CallbackEmpty},
		{Text: weekDayNames[4], CallbackData: types.CallbackEmpty},
		{Text: weekDayNames[5], CallbackData: types.CallbackEmpty},
		{Text: weekDayNames[6], CallbackData: types.CallbackEmpty},
	})
	//time.Utc заменить на user.Timezone
	fistMonthWeekday := time.Date(chosenYear, chosenMonth, 1, 0, 0, 0, 0, time.UTC).Weekday()
	daysInMonth := daysIn(chosenMonth, chosenYear)
	var weekButtons []gotgbot.InlineKeyboardButton
	var i time.Weekday
	if fistMonthWeekday > 0 {
		for i = 1; i < fistMonthWeekday; i++ {
			weekButtons = append(weekButtons, gotgbot.InlineKeyboardButton{
				Text:         " ",
				CallbackData: types.CallbackEmpty,
			})
		}
	} else {
		for i = 0; i < 6; i++ {
			weekButtons = append(weekButtons, gotgbot.InlineKeyboardButton{
				Text:         " ",
				CallbackData: types.CallbackEmpty,
			})
		}
	}
	for i := 1; i <= daysInMonth; i++ {
		weekButtons = append(weekButtons, gotgbot.InlineKeyboardButton{
			Text:         fmt.Sprintf("%d", i),
			CallbackData: fmt.Sprintf("%sday:%d-month:%d-year:%d", types.CallbackDeadlineChoose, i, chosenMonth, chosenYear),
		})
		if len(weekButtons) == 7 {
			buttons = append(buttons, weekButtons)
			weekButtons = nil
		}
		if i == daysInMonth && len(weekButtons) < 7 && len(weekButtons) > 0 {
			missing := 7 - len(weekButtons)
			for j := 0; j < missing; j++ {
				weekButtons = append(weekButtons, gotgbot.InlineKeyboardButton{
					Text:         " ",
					CallbackData: types.CallbackEmpty,
				})
			}
			buttons = append(buttons, weekButtons)
		}
	}
	return buttons
}

func CreateYearsButtons(chosenYear int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	var yearRow []gotgbot.InlineKeyboardButton
	for i := chosenYear - 7; i < chosenYear-2; i++ {
		yearRow = append(yearRow, gotgbot.InlineKeyboardButton{
			Text:         fmt.Sprintf("%d", i),
			CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseYear, i),
		})
	}
	buttons = append(buttons, yearRow)
	yearRow = nil
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("%d", chosenYear-2), CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseYear, chosenYear-2)},
		{Text: fmt.Sprintf("%d", chosenYear-1), CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseYear, chosenYear-1)},
		{Text: fmt.Sprintf("%d", chosenYear), CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseYear, chosenYear)},
		{Text: fmt.Sprintf("%d", chosenYear+1), CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseYear, chosenYear+1)},
		{Text: fmt.Sprintf("%d", chosenYear+2), CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseYear, chosenYear+2)},
	})
	for i := chosenYear + 3; i <= chosenYear+7; i++ {
		yearRow = append(yearRow, gotgbot.InlineKeyboardButton{
			Text:         fmt.Sprintf("%d", i),
			CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseYear, i),
		})
	}
	buttons = append(buttons, yearRow)
	return buttons
}

func CreateMonthsButtons(chosenMonth int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Январь", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 1)},
		{Text: "Февраль", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 2)},
		{Text: "Март", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 3)},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Апрель", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 4)},
		{Text: "Май", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 5)},
		{Text: "Июнь", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 6)},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Июль", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 7)},
		{Text: "Август", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 8)},
		{Text: "Сентябрь", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 9)},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Октябрь", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 10)},
		{Text: "Ноябрь", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 11)},
		{Text: "Декабрь", CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, 12)},
	})
	return buttons
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
