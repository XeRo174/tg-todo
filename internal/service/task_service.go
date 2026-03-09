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

// TaskMessageFill - формирует сообщение работы с задачей
func TaskMessageFill(title, ending string, task types.TaskModel, themes []types.ThemeModel) string {
	return fmt.Sprintf("%s\nНазвание: %s\nПриоритет: %s\nСроки: %s\nТемы: %s\n\n%s", title, task.Name, task.Priority.String(), task.Deadline.Format(types.TimeLayout), utils.ThemeStroke(themes), ending)
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

// CallbackTaskDone - обработчик обратного вызова завершения работы с задачей
func (s *Service) CallbackTaskDone(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Завершение работы с задачей"}); err != nil {
		return fmt.Errorf("ответ на завершение работы с задачей: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение создания задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		task.Status = types.TaskStatusCreated
		if err = s.Repository.UpdateTask(task); err != nil {
			return fmt.Errorf("обновление статуса задачи: %w", err)
		}
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	message := TaskMessageFill(fmt.Sprintf("%s завершено", operationName), "", task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, '%s' завершено: %w", operationName, err)
	}
	if err = s.Repository.DeleteMessageRegisterByUserId(user.ID); err != nil {
		return fmt.Errorf("очиска сообщения %s: %w", operationName, err)
	}
	return handlers.EndConversation()
}

// CallbackTaskFieldSkip - обработчик обратного вызова пропуска поля
func (s *Service) CallbackTaskFieldSkip(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Поле пропущено"}); err != nil {
		return fmt.Errorf("ответ на пропуск поля: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение создания задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	nextField := strings.Replace(callQuery.Data, types.CallbackTaskFieldSkip, "", 1)
	if nextField == "" {
		return nil
	}
	switch nextField {
	case types.ConversationTaskCreateSetName:
		if err = s.GenerateTaskNameMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, "Создание задачи", types.ConversationTaskCreateSetPriority); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetName)
	case types.ConversationTaskEditSetName:
		if err = s.GenerateTaskNameMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, "Редактирование задачи", types.ConversationTaskEditSetPriority); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskEditSetName)
	case types.ConversationTaskCreateSetPriority:
		if err = s.GenerateTaskPriorityMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, "Создание задачи", types.ConversationTaskCreateSetDeadline); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetPriority)
	case types.ConversationTaskEditSetPriority:
		if err = s.GenerateTaskPriorityMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, "Редактирование задачи", types.ConversationTaskEditSetDeadline); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskEditSetPriority)
	case types.ConversationTaskCreateSetDeadline:
		if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, "Создание задачи", types.ConversationTaskCreateSetTheme); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetDeadline)
	case types.ConversationTaskEditSetDeadline:
		if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, "Редактирование задачи", types.ConversationTaskEditSetTheme); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskEditSetDeadline)
	case types.ConversationTaskCreateSetTheme:
		if err = s.GenerateTaskThemeMessage(b, ctx.EffectiveSender.ChatId, userTGId, messageRegister, task, "Создание задачи", ""); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetTheme)
	case types.ConversationTaskEditSetTheme:
		if err = s.GenerateTaskThemeMessage(b, ctx.EffectiveSender.ChatId, userTGId, messageRegister, task, "Редактирование задачи", ""); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskEditSetTheme)
	default:
		return nil
	}
}

// CallbackTaskCancel - обработчик обратного вызова отмены работы с задачей
func (s *Service) CallbackTaskCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Отмена работы с задачей"}); err != nil {
		return fmt.Errorf("ответ на отмену работы с задачей: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение создания задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		task.Status = types.TaskStatusDropped
		if err = s.Repository.UpdateTask(task); err != nil {
			return fmt.Errorf("обновление статуса задачи: %w", err)
		}
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	message := TaskMessageFill(fmt.Sprintf("%s прекращено", operationName), "", task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, '%s' прекращено: %w", operationName, err)
	}
	if err = s.Repository.DeleteMessageRegisterByUserId(user.ID); err != nil {
		return fmt.Errorf("очиска сообщения %s: %w", operationName, err)
	}
	return handlers.EndConversation()
}

// CallbackTaskDoneTheme - обработчик обратного вызова завершения выбора тем
func (s *Service) CallbackTaskDoneTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Выбор завершен"}); err != nil {
		return fmt.Errorf("ответ на завершение выбора: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение создания задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	message := TaskMessageFill(fmt.Sprintf("%s", operationName), "Нажмите Завершить", task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TaskInlineKeyboard("", false),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задача, '%s' выбор тем завершен: %w", operationName, err)
	}
	return nil
}

// CallbackTaskChangeThemesPage - обработчик обратного вызова смены страницы, выбора темы задачи
func (s *Service) CallbackTaskChangeThemesPage(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Страница тем изменена"}); err != nil {
		return fmt.Errorf("ответ на переключение страницы: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	pageStr := strings.Replace(callQuery.Data, types.CallbackChangePage, "", 1)
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fmt.Errorf("получение номера страницы клавиатуры темы: %w", err)
	}
	themeFilter := types.ThemeFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: uint(page)}}
	themes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем для клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	message := TaskMessageFill(fmt.Sprintf("%s", operationName), fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeForTaskInlineKeyboard(task, themes, int(themesPagesCount), page),
		},
	}); err != nil {
		return fmt.Errorf("изменение страницы клавиатуры тем: %w", err)
	}
	return nil

}

// CallbackTaskChangeTasksPage - обработчик обратного вызова перехода на следующую страницу клавиатуры задач
func (s *Service) CallbackTaskChangeTasksPage(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Страница задач изменена"}); err != nil {
		return fmt.Errorf("ответ на переключение страницы: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	pageStr := strings.Replace(callQuery.Data, types.CallbackChangePage, "", 1)
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fmt.Errorf("получение номера страницы клавиатуры задач: %w", err)
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
	message := fmt.Sprintf("Выберите задачу для редактирования.\n\nСтраница: (%d/%d)", page, int(tasksPagesCount))
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: user.Messages[0].BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseTaskInlineKeyboard(tasks, page, int(tasksPagesCount)),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, смена страницы клавиатуры задач: %w", err)
	}
	return nil
}

// GenerateTaskNameMessage - сформировать сообщение запроса имени задачи
func (s *Service) GenerateTaskNameMessage(b *gotgbot.Bot, chatId int64, messageRegister types.MessageRegisterModel, task types.TaskModel, messageTitle, stage string) error {
	message := TaskMessageFill(messageTitle, "Введите имя задачи", task, task.Themes)
	if _, _, err := b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    chatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TaskInlineKeyboard(stage, true),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи: %w", err)
	}
	return nil
}

// GenerateTaskPriorityMessage - сформировать сообщение запроса приоритета задачи
func (s *Service) GenerateTaskPriorityMessage(b *gotgbot.Bot, chatId int64, messageRegister types.MessageRegisterModel, task types.TaskModel, messageTitle, stage string) error {
	message := TaskMessageFill(messageTitle, "Выберите приоритет", task, task.Themes)
	if _, _, err := b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    chatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.PriorityButtons(), utils.TaskInlineKeyboard(stage, true)...),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи: %w", err)
	}
	return nil
}

// GenerateTaskDeadlineMessage - сформировать сообщение запроса сроков задачи
func (s *Service) GenerateTaskDeadlineMessage(b *gotgbot.Bot, chatId int64, messageRegister types.MessageRegisterModel, task types.TaskModel, messageTitle, stage string) error {
	message := TaskMessageFill(messageTitle, fmt.Sprintf("Введите срок в формате %s", types.TimeLayout), task, task.Themes)
	if _, _, err := b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    chatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.CreateCalendarButtons(), utils.TaskInlineKeyboard(stage, true)...),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, приоритет задачи: %w", err)
	}
	return nil
}

// GenerateTaskThemeMessage - сформировать сообщение запроса тем задачи
func (s *Service) GenerateTaskThemeMessage(b *gotgbot.Bot, chatId, userTGId int64, messageRegister types.MessageRegisterModel, task types.TaskModel, messageTitle, stage string) error {
	themeFilter := types.ThemeFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: 1}}
	themes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	message := TaskMessageFill(messageTitle, fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    chatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeForTaskInlineKeyboard(task, themes, int(themesPagesCount), 1),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, сроки задачи :%w", err)
	}
	return nil
}
