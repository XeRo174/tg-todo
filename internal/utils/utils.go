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

func TimezonesInlineKeyboard() [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Калиниград (UTC+2)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Europe/Kaliningrad")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Москва (UTC+3)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Europe/Moscow")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Самара (UTC+4)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Europe/Samara")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Екатеринбург (UTC+5)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Asia/Yekaterinburg")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Омск (UTC+6)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Asia/Omsk")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Красноярск (UTC+7)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Asia/Krasnoyarsk")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Иркутск (UTC+8)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Asia/Irkutsk")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Якутстк (UTC+9)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Asia/Yakutsk")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Владивосток (UTC+10)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Asia/Vladivostok")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Магадан (UTC+11)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Asia/Magadan")},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("Камчатка (UTC+12)"), CallbackData: fmt.Sprintf("%s%s", types.CallbackUserSetTimezone, "Asia/Kamchatka")},
	})
	return buttons
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
func CreateCalendarButtons(chosenDate time.Time) [][]gotgbot.InlineKeyboardButton {

	chosenYear := chosenDate.Year()
	chosenMonth := chosenDate.Month()
	chosenDay := chosenDate.Day()
	chosenHour := chosenDate.Hour()
	chosenMinute := chosenDate.Minute()
	var buttons [][]gotgbot.InlineKeyboardButton
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("%d", chosenYear), CallbackData: fmt.Sprintf("%s%s", types.CallbackDeadlineShow, types.DeadlineShowYears)},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: chosenMonth.String(), CallbackData: fmt.Sprintf("%s%s", types.CallbackDeadlineShow, types.DeadlineShowMonths)},
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
		var dayStr string
		if i == chosenDay {
			dayStr = fmt.Sprintf("~%d~", i)
		} else {
			dayStr = fmt.Sprintf("%d", i)
		}
		weekButtons = append(weekButtons, gotgbot.InlineKeyboardButton{
			Text:         dayStr,
			CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseDay, i),
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
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: fmt.Sprintf("%d", chosenHour), CallbackData: fmt.Sprintf("%s%s", types.CallbackDeadlineShow, types.DeadlineShowHours)},
		{Text: fmt.Sprintf("%d", chosenMinute), CallbackData: fmt.Sprintf("%s%s", types.CallbackDeadlineShow, types.DeadlineShowMinutes)},
	})
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Завершить установку сроков", CallbackData: types.CallbackTaskSetDeadlineDone},
	})
	return buttons
}

func CreateYearsButtons(chosenYear int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	var yearRow []gotgbot.InlineKeyboardButton

	for i := chosenYear - 7; i <= chosenYear+7; i++ {
		var yearStr string
		if i == chosenYear {
			yearStr = fmt.Sprintf("~%d~", i)
		} else {
			yearStr = fmt.Sprintf("%d", i)
		}
		yearRow = append(yearRow, gotgbot.InlineKeyboardButton{
			Text:         yearStr,
			CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseYear, i),
		})
		if len(yearRow) == 5 {
			buttons = append(buttons, yearRow)
			yearRow = nil
		}
	}
	return buttons
}

func CreateMonthsButtons(chosenMonth int) [][]gotgbot.InlineKeyboardButton {
	type MonthStruct struct {
		Name   string
		Number int
	}
	months := []MonthStruct{
		{Name: "Январь", Number: 1},
		{Name: "Февраль", Number: 2},
		{Name: "Март", Number: 3},
		{Name: "Апрель", Number: 4},
		{Name: "Май", Number: 5},
		{Name: "Июнь", Number: 6},
		{Name: "Июль", Number: 7},
		{Name: "Август", Number: 8},
		{Name: "Сентябрь", Number: 9},
		{Name: "Октябрь", Number: 10},
		{Name: "Ноябрь", Number: 11},
		{Name: "Декабрь", Number: 12},
	}
	var buttons [][]gotgbot.InlineKeyboardButton
	var buttonRow []gotgbot.InlineKeyboardButton
	for _, month := range months {
		var monthStr string
		if month.Number == chosenMonth {
			monthStr = fmt.Sprintf("~%s~", month.Name)
		} else {
			monthStr = fmt.Sprintf("%s", month.Name)
		}
		buttonRow = append(buttonRow, gotgbot.InlineKeyboardButton{
			Text: monthStr, CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMonth, month.Number),
		})
		if len(buttonRow) == 3 {
			buttons = append(buttons, buttonRow)
			buttonRow = nil
		}
	}
	return buttons
}

func CreateHoursButtons(chosenHour int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	var buttonRow []gotgbot.InlineKeyboardButton
	for i := 0; i < 25; i++ {
		var hourStr string
		if i == chosenHour {
			hourStr = fmt.Sprintf("~%d~", i)
		} else {
			hourStr = fmt.Sprintf("%d", i)
		}
		buttonRow = append(buttonRow, gotgbot.InlineKeyboardButton{
			Text:         hourStr,
			CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseHour, i),
		})
		if len(buttonRow) == 4 {
			buttons = append(buttons, buttonRow)
			buttonRow = nil
		}
	}
	return buttons
}

func CreateMinutesButtons(chosenMinute int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	var buttonRow []gotgbot.InlineKeyboardButton
	for i := 0; i < 60; i++ {
		var minuteStr string
		if i == chosenMinute {
			minuteStr = fmt.Sprintf("~%d~", i)
		} else {
			minuteStr = fmt.Sprintf("%d", i)
		}
		buttonRow = append(buttonRow, gotgbot.InlineKeyboardButton{
			Text:         minuteStr,
			CallbackData: fmt.Sprintf("%s%d", types.CallbackDeadlineChooseMinute, i),
		})
		if len(buttonRow) == 6 {
			buttons = append(buttons, buttonRow)
			buttonRow = nil
		}
	}
	return buttons
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
