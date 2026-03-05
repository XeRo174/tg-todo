package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"strconv"
	"strings"
	"tg-todo/internal/types"
	"tg-todo/internal/utils"
)

func TaskMessageFill(title, ending string, task types.TaskModel, themes []types.ThemeModel) string {
	return fmt.Sprintf("%s\nНазвание: %s\nПриоритет: %s\nСроки: %s\nТемы: %s\n\n%s", title, task.Name, task.Priority.String(), task.Deadline.Format(types.TimeLayout), utils.ThemeStroke(themes), ending)
}

// ConversationCreateTaskCancel - отмена создания
func (s *Service) ConversationCreateTaskCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Отмена создания"}); err != nil {
		return fmt.Errorf("ответ на отмену создания: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней новой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationCreate)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	lastTask.Status = types.TaskStatusDropped
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление статуса задачи: %w", err)
	}
	message := TaskMessageFill("Задача отменена", "", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, отмена создания: %w", err)
	}
	return handlers.EndConversation()
}

func (s *Service) ConversationCreateTaskChangePage(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Страница тем изменена"}); err != nil {
		return fmt.Errorf("ответ на переключение страницы: %w", err)
	}
	pageStr := strings.Replace(callQuery.Data, types.CallbackChangeTaskThemesPage, "", 1)
	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		return fmt.Errorf("получение номера страницы клавиатуры тем: %w", err)
	}
	themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: uint(page)}}
	allThemes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получения тем для клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней новой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationCreate)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, int(themesPagesCount)), lastTask, lastTask.Themes)
	if _, _, err = callQuery.Message.EditText(b, message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.CreateThemeInlineKeyboard(lastTask, allThemes, uint(page), themesPagesCount),
		},
	}); err != nil {
		return fmt.Errorf("изменение страницы клавиатуры тем: %w", err)
	}
	return nil
}

func (s *Service) ConversationCreateTaskSkipField(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Поле пропущено"}); err != nil {
		return fmt.Errorf("ответ на пропуск поля: %w", err)
	}
	nextField := strings.Replace(callQuery.Data, types.CallbackTaskFieldSkip, "", 1)
	if nextField == "" {
		return nil
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней новой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationCreate)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	switch nextField {
	case types.ConversationTaskCreateSetName:
		message := TaskMessageFill("Создание задачи", "Введите имя задачи", lastTask, lastTask.Themes)
		if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
			MessageId: messageRegister.BotMessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: append(utils.PriorityButtons(), utils.TaskButtons(types.ConversationTaskCreateSetPriority)...),
			},
		}); err != nil {
			return fmt.Errorf("изменение сообщения задачи: %w", err)
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetName)
	case types.ConversationTaskCreateSetPriority:
		message := TaskMessageFill("Создание задачи", "Выберите приоритет", lastTask, lastTask.Themes)
		if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
			MessageId: messageRegister.BotMessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: append(utils.PriorityButtons(), utils.TaskButtons(types.ConversationTaskCreateSetDeadline)...),
			},
		}); err != nil {
			return fmt.Errorf("изменение сообщения задачи: %w", err)
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetPriority)
	case types.ConversationTaskCreateSetDeadline:
		message := TaskMessageFill("Создание задачи", fmt.Sprintf("Введите срок в формате %s", types.TimeLayout), lastTask, lastTask.Themes)
		if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
			MessageId: messageRegister.BotMessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: utils.TaskButtons(types.ConversationTaskCreateSetTheme),
			},
			//todo inline keyboard календарь + время
		}); err != nil {
			return fmt.Errorf("изменение сообщения задачи, приоритет задачи: %w", err)
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetDeadline)
	case types.ConversationTaskCreateSetTheme:
		themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: 1}}
		themes, err := s.Repository.GetThemes(themeFilter)
		if err != nil {
			return fmt.Errorf("получение тем для клавиатуры: %w", err)
		}
		themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
		if err != nil {
			return fmt.Errorf("получение количества страниц тем: %w", err)
		}
		message := TaskMessageFill("Создание задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), lastTask, lastTask.Themes)
		if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
			MessageId: messageRegister.BotMessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: s.CreateThemeInlineKeyboard(lastTask, themes, 1, themesPagesCount),
			},
		}); err != nil {
			return fmt.Errorf("изменение сообщения задачи, сроки задачи: %w", err)
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetTheme)
	default:
		return nil
	}
}

// CommandGetTasks - обработчик команды получения задач
func (s *Service) CommandGetTasks(b *gotgbot.Bot, ctx *ext.Context) error {
	tasks, err := s.Repository.GetTasks(types.TaskFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.UnlimitedSize, Page: 1}})
	if err != nil {
		return fmt.Errorf("получение задач: %w", err)
	}
	var taskStroke []string
	for _, task := range tasks {
		var themeStroke []string
		for _, theme := range task.Themes {
			themeStroke = append(themeStroke, fmt.Sprintf("%s", theme.Name))
		}
		priority, ok := types.ParsePriority(fmt.Sprintf("%d", task.Priority))
		if !ok {
			priority = types.TaskPriorityNone
		}
		stroke := fmt.Sprintf("Задача: %s. Приоритет: %s. Статус: %s. Срок %s. Темы: %s.", task.Name, priority.String(), task.Status.String(), task.Deadline.Format(types.TimeLayout), strings.Join(themeStroke, ", "))
		taskStroke = append(taskStroke, stroke)
	}
	var message string
	if len(taskStroke) > 0 {
		message = strings.Join(taskStroke, "\n\n")
	} else {
		message = "Нет задач"
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, message, nil); err != nil {
		return fmt.Errorf("отправка задач: %w", err)
	}
	return nil
}

func (s *Service) CreateThemeInlineKeyboard(lastTask types.TaskModel, themesByPage []types.ThemeModel, page uint, pagesCount float64) [][]gotgbot.InlineKeyboardButton {
	//todo переделать под новый формат
	var buttons [][]gotgbot.InlineKeyboardButton
	var chosenThemes []types.ThemeModel
	for _, theme := range themesByPage {
		var buttonText string
		if utils.Contains(lastTask.Themes, theme.Name) {
			buttonText = fmt.Sprintf("[x] %s", theme.Name)
			chosenThemes = append(chosenThemes, theme)
		} else {
			buttonText = fmt.Sprintf("[  ] %s", theme.Name)
		}
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: buttonText, CallbackData: fmt.Sprintf("%s%d;%s%d", types.CallbackTaskThemeChoose, theme.ID, types.CallbackCurrentPage, page)},
		})
	}
	totalPages := int(pagesCount)
	currentPage := int(page)

	if totalPages > 1 {
		if currentPage == 1 {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", totalPages, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, totalPages)},
				{Text: fmt.Sprintf("(%d/%d) >", currentPage+1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage+1)},
			})
		} else if currentPage == totalPages {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", 1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, 1)},
			})
		} else if currentPage == totalPages {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", 1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, 1)},
			})
		} else {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", currentPage+1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage+1)},
			})
		}
	}
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Завершить", CallbackData: types.CallbackThemeChoseDone},
	})
	return append(buttons, utils.TaskButtons("")...)
}

// Новый формат функций

// ChooseThemeForTaskInlineKeyboard - создает клавиатуру для выбора тем задачи с пагинацией
func (s *Service) ChooseThemeForTaskInlineKeyboard(task types.TaskModel, themesByPage []types.ThemeModel, totalPages, currentPage int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, theme := range themesByPage {
		if utils.Contains(task.Themes, theme.Name) {
			buttonText := fmt.Sprintf("[x] %s", theme.Name)
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				//todo Callback для удаления темы и нормальное название, может select и unselect
				{Text: buttonText, CallbackData: fmt.Sprintf("%s%d;%s%d", types.CallbackTaskThemeUnChoose, theme.ID, types.CallbackCurrentPage, currentPage)},
			})
		} else {
			buttonText := fmt.Sprintf("[  ] %s", theme.Name)
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: buttonText, CallbackData: fmt.Sprintf("%s%d;%s%d", types.CallbackTaskThemeChoose, theme.ID, types.CallbackCurrentPage, currentPage)},
			})
		}
	}
	if totalPages > 1 {
		buttons = append(buttons, utils.CreateArrowButtons(currentPage, totalPages)...) //todo передача callback для переключения страниц возможно
	}
	return buttons
}

// TaskInlineKeyboard - создает клавиатуру работы с задачей
func (s *Service) TaskInlineKeyboard(nextField string) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	//todo не Done и Cancel, а Complete(d) и Stop(ed)
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Done", CallbackData: types.CallbackTaskDone},
		{Text: "Skip", CallbackData: fmt.Sprintf("%s%s", types.CallbackTaskFieldSkip, nextField)},
		{Text: "Cancel", CallbackData: types.CallbackTaskCancel},
	})
	return buttons
}

// ChooseTaskInlineKeyboard - создает клавиатуру выбора задач с пагинацией
func (s *Service) ChooseTaskInlineKeyboard(tasksByPage []types.TaskModel, totalPages, currentPage int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, task := range tasksByPage {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: task.Name, CallbackData: fmt.Sprintf("%s%d", types.CallbackTaskChoose, task.ID)},
		})
	}
	if totalPages > 1 {
		buttons = append(buttons, utils.CreateArrowButtons(currentPage, totalPages)...)
	}
	return buttons
}

//todo что если вместо CallbackTaskCancel сделать CallbackCreateTaskCancel и CallbackEditTaskCancel, тогда не нужно будет строить условия и передавать тип операции внутри callback

// CallbackTaskCancel - обработчик обратного вызова отмены работы с задачей
func (s *Service) CallbackTaskCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Отмена работы с задачей"}); err != nil {
		return fmt.Errorf("ответ на отмену работы с задачей: %w", err)
	}
	//Получаем тип операции, который проводился с задачей до этого
	canceledOperation := strings.Replace(callQuery.Data, types.CallbackTaskCancel, "", 1)
	//Получаем задачу с которой была работа
	var lastTask types.TaskModel
	var err error
	var operationName string
	switch canceledOperation {
	case types.MessageRegisterOperationEdit:
		lastTask, err = s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationCreate:
		lastTask, err = s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", canceledOperation)
	}
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	//Получаем сообщение задачи
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, canceledOperation)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	//Если мы создавали задачу, а не редактировали, то меняем статус на Брошенная.
	// todo Нужно ли менять статус на Брошенная, статус черновик и так удовлетворяет требованиям
	if canceledOperation == types.MessageRegisterOperationCreate {
		lastTask.Status = types.TaskStatusDropped
		if err = s.Repository.UpdateTask(lastTask); err != nil {
			return fmt.Errorf("обновление статуса задачи: %w", err)
		}
	}
	//Меняем сообщение, с указанием, что работа с задачей отменена
	// todo Отмена? это ведь значит, что выполненные действия нужно откатить, разве нет? может тогда переделать на Прекращены и на Stop вместо Cancel?
	message := TaskMessageFill(fmt.Sprintf("%s отменено", operationName), "", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения '%s', отмена работы с задачей: %w", operationName, err)
	}
	return handlers.EndConversation()
}

// CallbackTaskDone - обработчик обратного вызова завершения работы с задачей
func (s *Service) CallbackTaskDone(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Отмена работы с задачей"}); err != nil {
		return fmt.Errorf("ответ на отмену работы с задачей: %w", err)
	}
	//Получаем тип операции, который проводился с задачей до этого
	completedOperation := strings.Replace(callQuery.Data, types.CallbackTaskComplete, "", 1)
	var lastTask types.TaskModel
	var err error
	var operationName string
	switch completedOperation {
	case types.MessageRegisterOperationEdit:
		lastTask, err = s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationCreate:
		lastTask, err = s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", completedOperation)
	}
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, completedOperation)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	if completedOperation == types.MessageRegisterOperationCreate {
		lastTask.Status = types.TaskStatusCreated
		if err = s.Repository.UpdateTask(lastTask); err != nil {
			return fmt.Errorf("обновление статуса задачи: %w", err)
		}
	}
	message := TaskMessageFill(fmt.Sprintf("%s завершено", operationName), "", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения '%s', завершение работы с задачей: %w", operationName, err)
	}
	return handlers.EndConversation()
}

// CallbackTaskChangeTasksPage - обработчик обратного вызова перехода на следующую страницу клавиатуры задач
func (s *Service) CallbackTaskChangeTasksPage(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Изменение страницы задач"}); err != nil {
		return fmt.Errorf("ответ на смену страницы задач: %w", err)
	}
	pageStr := strings.Replace(callQuery.Data, types.CallbackChangePage, "", 1)
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fmt.Errorf("получение номера страницы клавиатуры задач: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение последней редактируемой задачи не найдено")
	}
	taskFilter := types.TaskFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: uint(page)}}
	tasks, err := s.Repository.GetTasks(taskFilter)
	if err != nil {
		return fmt.Errorf("получение задач для клавиатуры: %w", err)
	}
	tasksPagesCount, err := s.Repository.GetTaskPages(taskFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц задач: %w", err)
	}
	message := fmt.Sprintf("Выберите задачу.\n\nСтраница: (%d/%d)", page, int(tasksPagesCount))
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.ChooseTaskInlineKeyboard(tasks, page, int(tasksPagesCount)),
		},
	}); err != nil {
		return fmt.Errorf("изменение страницы клавиатуры задач :%w", err)
	}
	return nil
}
