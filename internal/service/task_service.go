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
	"time"
)

// TaskMessageFill - формирует сообщение работы с задачей
func TaskMessageFill(title, ending string, task types.TaskModel, themes []types.ThemeModel) string {
	return fmt.Sprintf("%s\nНазвание: %s\nПриоритет: %s\nСроки: %s\nТемы: %s\n\n%s", title, task.Name, task.Priority.String(), task.Deadline.Format(types.TimeLayout), utils.ThemeStroke(themes), ending)
}

// ConversationTasksInit - обработчик разговора инициализации работы с задачей
func (s *Service) ConversationTasksInit(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) > 0 {
		messageRegister := user.Messages[0]
		if _, _, err = b.EditMessageText(MessageOperationBeauty(messageRegister), &gotgbot.EditMessageTextOpts{MessageId: messageRegister.BotMessageId, ChatId: ctx.EffectiveSender.ChatId}); err != nil {
			s.App.Logger.Warn(fmt.Errorf("не удалось изменить прошлое сообщение: %w", err).Error())
		}
	}
	taskFilter := types.TaskFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: 1}}
	tasks, err := s.Repository.GetTasks(taskFilter)
	if err != nil {
		return fmt.Errorf("получение задач для выбора: %w", err)
	}
	tasksPagesCount, err := s.Repository.GetTaskPages(taskFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц задач для выбора: %w", err)
	}
	message := fmt.Sprintf("Ваши задачи.\n\nСтраница: (1/%d)", int(tasksPagesCount))
	taskMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseTaskInlineKeyboard(tasks, int(tasksPagesCount), 1),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения задач: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(types.MessageRegisterModel{
		UserId:       user.ID,
		BotMessageId: taskMessage.MessageId,
		Operation:    types.MessageRegisterOperationTask,
	}); err != nil {
		return fmt.Errorf("запись сообщения списка задач: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskChoose)
}

// ConversationTaskChoose - обработчик разговора выбора задачи для работы
func (s *Service) ConversationTaskChoose(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Задача выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор задачи: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	callQueryValues := strings.Split(callQuery.Data, ";")
	var taskId, currentPage uint
	for _, callQueryData := range callQueryValues {
		if strings.HasPrefix(callQueryData, types.CallbackTaskChoose) {
			taskStr := strings.Replace(callQueryData, types.CallbackTaskChoose, "", 1)
			task, err := strconv.Atoi(taskStr)
			if err != nil {
				return fmt.Errorf("получение идентификатора задачи из клавиатуры: %w", err)
			}
			taskId = uint(task)
		}
		if strings.HasPrefix(callQueryData, types.CallbackCurrentPage) {
			pageStr := strings.Replace(callQueryData, types.CallbackCurrentPage, "", 1)
			page, err := strconv.ParseUint(pageStr, 10, 64)
			if err != nil {
				return fmt.Errorf("получение номера страницы клавиатуры тем: %w", err)
			}
			currentPage = uint(page)
		}
	}
	task, err := s.Repository.GetTaskById(ctx.EffectiveSender.User.Id, taskId)
	if err != nil {
		return fmt.Errorf("выбранная задача не найдена: %w", err)
	}
	messageRegister := user.Messages[0]
	messageRegister.TaskId = task.ID
	message := TaskMessageFill("Информация о задаче", "Выберите действие", task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TaskActionButtons(currentPage),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(messageRegister); err != nil {
		return fmt.Errorf("запись сообщения задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskActionChoose)
}

// ConversationChooseTaskAction - обработчик разговора выбора действия с задачей
func (s *Service) ConversationChooseTaskAction(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Действие выбрано"}); err != nil {
		return fmt.Errorf("ответ на выбор действий: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	callQueryValues := strings.Split(callQuery.Data, ";")
	var taskAction string
	var currentPage uint
	for _, callQueryData := range callQueryValues {
		if strings.HasPrefix(callQueryData, types.CallbackTaskAction) {
			taskAction = strings.Replace(callQueryData, types.CallbackTaskAction, "", 1)
		}
		if strings.HasPrefix(callQueryData, types.CallbackCurrentPage) {
			pageStr := strings.Replace(callQueryData, types.CallbackCurrentPage, "", 1)
			page, err := strconv.ParseUint(pageStr, 10, 64)
			if err != nil {
				return fmt.Errorf("получение номера страницы клавиатуры задач: %w", err)
			}
			currentPage = uint(page)
		}
	}
	var message, conversationState string
	var keyboard [][]gotgbot.InlineKeyboardButton
	switch taskAction {
	case types.ActionEdit:
		message = TaskMessageFill("Редактирование задачи", "Введите новое название задачи", task, task.Themes)
		conversationState = types.ConversationTaskEditSetName
		messageRegister.Operation = types.MessageRegisterOperationTaskEdit
		keyboard = utils.TaskInlineKeyboard(types.ConversationTaskEditSetPriority, true)
	case types.ActionStatus:
		message = TaskMessageFill("Установка статуса задачи", "Выберите статус задачи", task, task.Themes)
		conversationState = types.ConversationTaskSetStatus
		messageRegister.Operation = types.MessageRegisterOperationTaskStatus
		keyboard = utils.StatusButtons()
	case types.ActionDelete:
		message = TaskMessageFill("Удаление задачи", "Вы точно хотите удалить задачу?", task, task.Themes)
		conversationState = types.ConversationTaskDelete
		messageRegister.Operation = types.MessageRegisterOperationTaskDelete
	case types.ActionBack:
		taskFilter := types.TaskFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: currentPage}}
		tasks, err := s.Repository.GetTasks(taskFilter)
		if err != nil {
			return fmt.Errorf("получение задач для выбора: %w", err)
		}
		tasksPagesCount, err := s.Repository.GetTaskPages(taskFilter)
		if err != nil {
			return fmt.Errorf("получение количества страниц задач для выбора: %w", err)
		}
		message = fmt.Sprintf("Ваши задачи.\n\nСтраница: (%d/%d)", currentPage, int(tasksPagesCount))
		conversationState = types.ConversationTaskChoose
		keyboard = utils.ChooseTaskInlineKeyboard(tasks, int(tasksPagesCount), int(currentPage))
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, редактирование: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(messageRegister); err != nil {
		return fmt.Errorf("запись сообщения '%s': %w", messageRegister.Operation, err)
	}
	return handlers.NextConversationState(conversationState)
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
	message := fmt.Sprintf("Ваши задачи.\n\nСтраница: (%d/%d)", page, int(tasksPagesCount))
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: user.Messages[0].BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseTaskInlineKeyboard(tasks, int(tasksPagesCount), page),
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
func (s *Service) GenerateTaskDeadlineMessage(b *gotgbot.Bot, chatId int64, messageRegister types.MessageRegisterModel, task types.TaskModel, chosenDate time.Time, messageTitle, stage string) error {
	message := TaskMessageFill(messageTitle, "Укажите срок задачи", task, task.Themes)
	if _, _, err := b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    chatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.CreateCalendarButtons(chosenDate), utils.TaskInlineKeyboard(stage, true)...),
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

// ConversationTaskStatusSet - обработчик разговора установки статуса задачи
func (s *Service) ConversationTaskStatusSet(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Статус выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор статуса: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	status, ok := types.ParseStatus(callQuery.Data)
	if !ok {
		status = types.TaskStatusDraft
	}
	task.Status = status
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление статуса задачи: %w", err)
	}
	if _, _, err = b.EditMessageText(fmt.Sprintf("Статус задачи %s обновлен на %s", task.Name, task.Status), &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, установка нового статуса: %w", err)
	}
	if err = s.Repository.DeleteMessageRegisterByUserId(user.ID); err != nil {
		return fmt.Errorf("очистка сообщений установки статуса задачи: %w", err)
	}
	return handlers.EndConversation()

}

// GenerateTaskStatusMessage - сформировать сообщение запроса статуса задачи
func (s *Service) GenerateTaskStatusMessage(b *gotgbot.Bot, chatId int64, messageRegister types.MessageRegisterModel, task types.TaskModel, messageTitle string) error {
	message := TaskMessageFill(messageTitle, "Выберите статус задачи", task, task.Themes)
	if _, _, err := b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    chatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.StatusButtons(),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи: %w", err)
	}
	return nil
}
