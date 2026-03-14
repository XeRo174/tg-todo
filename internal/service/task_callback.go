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

//Общие callback обработчики задач

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
		if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, task.Deadline, "Создание задачи", types.ConversationTaskCreateSetTheme); err != nil {
			return err
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetDeadline)
	case types.ConversationTaskEditSetDeadline:
		if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, task.Deadline, "Редактирование задачи", types.ConversationTaskEditSetTheme); err != nil {
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

// Callback обработчики работы со сроками

// CallbackDeadlineShowChoose - обработчик обратного вызова выбора параметра сроков
func (s *Service) CallbackDeadlineShowChoose(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Отображение годов"}); err != nil {
		return fmt.Errorf("ответ на отображение годов: %w", err)
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
	var operationName, message string
	var keyboard [][]gotgbot.InlineKeyboardButton
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	show := strings.Replace(callQuery.Data, types.CallbackDeadlineShow, "", 1)
	switch show {
	case types.DeadlineShowYears:
		message = TaskMessageFill(operationName, "Выберите год", task, task.Themes)
		keyboard = utils.CreateYearsButtons(task.Deadline.Year())
	case types.DeadlineShowMonths:
		message = TaskMessageFill(operationName, "Выберите месяца", task, task.Themes)
		keyboard = utils.CreateMonthsButtons(int(task.Deadline.Month()))
	case types.DeadlineShowDays:
		message = TaskMessageFill(operationName, "Выберите день", task, task.Themes)
		keyboard = utils.CreateCalendarButtons(task.Deadline)
	case types.DeadlineShowHours:
		message = TaskMessageFill(operationName, "Выберите час", task, task.Themes)
		keyboard = utils.CreateHoursButtons(task.Deadline.Hour())
	case types.DeadlineShowMinutes:
		message = TaskMessageFill(operationName, "Выберите минуты", task, task.Themes)
		keyboard = utils.CreateMinutesButtons(task.Deadline.Minute())
	default:
		return fmt.Errorf("идентификация объекта сроков для показа: %s", show)
	}
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщений задачи, сроки задачи: %w", err)
	}
	return nil
}

// CallbackTaskDeadlineChooseYear - обработчик обратного вызова выбора года
func (s *Service) CallbackTaskDeadlineChooseYear(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Год выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор года: %w", err)
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
	yearStr := strings.Replace(callQuery.Data, types.CallbackDeadlineChooseYear, "", 1)
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return fmt.Errorf("получение года из клавиатуры: %w", err)
	}
	//time.Utc заменить на user.Timezone
	loc, err := time.LoadLocation(user.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	if task.Deadline.IsZero() {
		task.Deadline = time.Date(year, time.Now().Month(), 1, 0, 0, 0, 0, loc)
	} else {
		task.Deadline = time.Date(year, task.Deadline.Month(), 1, task.Deadline.Hour(), task.Deadline.Minute(), 0, 0, loc)
	}
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, task.Deadline, operationName, types.ConversationTaskCreateSetTheme); err != nil {
		return err
	}
	return nil
}

// CallbackTaskDeadlineChooseMonth - обработчик обратного вызова выбора месяца
func (s *Service) CallbackTaskDeadlineChooseMonth(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Месяц выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор месяца: %w", err)
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
	monthStr := strings.Replace(callQuery.Data, types.CallbackDeadlineChooseMonth, "", 1)
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return fmt.Errorf("получение года из клавиатуры: %w", err)
	}
	//time.Utc заменить на user.Timezone
	loc, err := time.LoadLocation(user.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	if task.Deadline.IsZero() {
		task.Deadline = time.Date(time.Now().Year(), time.Month(month), 1, 0, 0, 0, 0, loc)
	} else {
		task.Deadline = time.Date(task.Deadline.Year(), time.Month(month), 1, task.Deadline.Hour(), task.Deadline.Minute(), task.Deadline.Minute(), 0, loc)
	}
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, task.Deadline, operationName, types.ConversationTaskCreateSetTheme); err != nil {
		return err
	}
	return nil
}

// CallbackTaskDeadlineChooseDay - обработчик обратного вызова выбора дня
func (s *Service) CallbackTaskDeadlineChooseDay(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Час выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор часа: %w", err)
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
	dayStr := strings.Replace(callQuery.Data, types.CallbackDeadlineChooseDay, "", 1)
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return fmt.Errorf("получение дня из клавиатуры: %w", err)
	}
	//time.Utc заменить на user.Timezone
	loc, err := time.LoadLocation(user.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	if task.Deadline.IsZero() {
		task.Deadline = time.Date(time.Now().Year(), time.Now().Month(), day, 0, 0, 0, 0, loc)
	} else {
		task.Deadline = time.Date(task.Deadline.Year(), task.Deadline.Month(), day, task.Deadline.Hour(), task.Deadline.Minute(), 0, 0, loc)
	}
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, task.Deadline, operationName, types.ConversationTaskCreateSetTheme); err != nil {
		return err
	}
	return nil
}

// CallbackTaskDeadlineChooseHour - обработчик обратного вызова выбора часа
func (s *Service) CallbackTaskDeadlineChooseHour(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Час выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор часа: %w", err)
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
	hourStr := strings.Replace(callQuery.Data, types.CallbackDeadlineChooseHour, "", 1)
	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return fmt.Errorf("получение часа из клавиатуры: %w", err)
	}
	//time.Utc заменить на user.Timezone
	loc, err := time.LoadLocation(user.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	if task.Deadline.IsZero() {
		task.Deadline = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hour, 0, 0, 0, loc)
	} else {
		task.Deadline = time.Date(task.Deadline.Year(), task.Deadline.Month(), task.Deadline.Day(), hour, task.Deadline.Minute(), 0, 0, loc)
	}
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, task.Deadline, operationName, types.ConversationTaskCreateSetTheme); err != nil {
		return err
	}
	return nil
}

// CallbackTaskDeadlineChooseMinute - обработчик обратного вызова выбора минут
func (s *Service) CallbackTaskDeadlineChooseMinute(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Минуты выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор минут: %w", err)
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
	minuteStr := strings.Replace(callQuery.Data, types.CallbackDeadlineChooseMinute, "", 1)
	minute, err := strconv.Atoi(minuteStr)
	if err != nil {
		return fmt.Errorf("получение минут из клавиатуры: %w", err)
	}
	//time.Utc заменить на user.Timezone
	loc, err := time.LoadLocation(user.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	if task.Deadline.IsZero() {
		task.Deadline = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), minute, 0, 0, loc)
	} else {
		task.Deadline = time.Date(task.Deadline.Year(), task.Deadline.Month(), task.Deadline.Day(), task.Deadline.Hour(), minute, 0, 0, loc)
	}
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	var operationName string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, task.Deadline, operationName, types.ConversationTaskCreateSetTheme); err != nil {
		return err
	}
	return nil
}

// CallbackTaskDoneDeadline - обработчик обратного вызова завершения установки сроков
func (s *Service) CallbackTaskDoneDeadline(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Установка сроков завершена"}); err != nil {
		return fmt.Errorf("ответ на завершение установки сроков: %w", err)
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
	var operationName, nextStage string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskEdit:
		operationName = "Редактирование задачи"
		nextStage = types.ConversationTaskEditSetTheme
	case types.MessageRegisterOperationTaskCreate:
		operationName = "Создание задачи"
		nextStage = types.ConversationTaskCreateSetTheme
	default:
		return fmt.Errorf("идентификация типа работы с задачей: %s", messageRegister.Operation)
	}
	if err = s.GenerateTaskThemeMessage(b, ctx.EffectiveSender.ChatId, userTGId, messageRegister, task, operationName, ""); err != nil {
		return err
	}
	return handlers.NextConversationState(nextStage)
}

// Callback обработчики работы с темами

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
	pageStr := strings.Replace(callQuery.Data, types.CallbackChangeThemeForTaskPage, "", 1)
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
