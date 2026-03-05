package service

import (
	"errors"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"tg-todo/internal/types"
	"tg-todo/internal/utils"
	"time"
)

func (s *Service) ConversationEditTaskInit(b *gotgbot.Bot, ctx *ext.Context) error {
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if ok {
		message := TaskMessageFill("Задача", "отмена прошлых действий", lastTask, lastTask.Themes)
		if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
			MessageId: messageRegister.BotMessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
		}); err != nil {
			return fmt.Errorf("изменение сообщения задачи: %w", err)
		}
	}
	taskFilter := types.TaskFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: 1}}
	tasks, err := s.Repository.GetTasks(taskFilter)
	if err != nil {
		return fmt.Errorf("получение задач для клавиатуры")
	}
	tasksPagesCount, err := s.Repository.GetTaskPages(taskFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц задач: %w", err)
	}
	message := fmt.Sprintf("Выберите задачу для редактирования.\n\nСтраница: (1/%d)", int(tasksPagesCount))
	taskMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.ChooseTaskInlineKeyboard(tasks, 1, int(tasksPagesCount)),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения редактирования задач: %w", err)
	}
	if err = s.Repository.WriteTaskMessage(types.MessageRegisterModel{TaskId: 0, BotMessageId: taskMessage.MessageId, Operation: types.MessageRegisterOperationEdit}); err != nil {
		return fmt.Errorf("запись сообщения редактирования задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskEditChooseTask)
}

func (s *Service) ConversationEditTaskChooseTask(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Задача выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор задачи: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	taskStr := strings.Replace(callQuery.Data, types.CallbackTaskChoose, "", 1)
	taskId, err := strconv.Atoi(taskStr)
	if err != nil {
		return fmt.Errorf("получение идентификатора задачи из клавиатуры: %w", err)
	}
	task, err := s.Repository.GetTaskById(ctx.EffectiveSender.User.Id, uint(taskId))
	if err != nil {
		return fmt.Errorf("выбранная задача не найдена: %w", err)
	}
	messageRegister.TaskId = task.ID
	if err = s.Repository.UpdateTaskMessageByMessageIdAndTaskId(messageRegister); err != nil {
		return fmt.Errorf("обновление регистра сообщений: %w", err)
	}
	message := TaskMessageFill("Редактирование задачи", "Введите новое название", task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщение редактирования задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetName)
}

func (s *Service) ConversationEditTaskChangeTasksPage(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Страница задач изменена"}); err != nil {
		return fmt.Errorf("ответ на переключение страницы: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok {
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
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.ChooseTaskInlineKeyboard(tasks, page, int(tasksPagesCount)),
		},
	}); err != nil {
		return fmt.Errorf("изменение страницы клавиатуры задач: %w", err)
	}
	return nil
}

func (s *Service) ConversationEditTaskSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	lastTask.Name = ctx.EffectiveMessage.Text
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление имени задачи: %w", err)
	}
	message := TaskMessageFill("Редактирование задачи", "Выберите приоритет", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.PriorityButtons(), s.TaskInlineKeyboard(types.ConversationTaskCreateSetDeadline)...),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения редактирования задачи: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаления сообщения нового имени задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetPriority)
}

func (s *Service) ConversationEditTaskSetPriority(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Приоритет выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор приоритета: %w", err)
	}
	priority, ok := types.ParsePriority(callQuery.Data)
	if !ok {
		priority = types.TaskPriorityNone
	}
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	lastTask.Priority = priority
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление приоритета задачи: %w", err)
	}
	message := TaskMessageFill("Редактирование задачи", fmt.Sprintf("Введите срок в формате %s", types.TimeLayout), lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.TaskInlineKeyboard(types.ConversationTaskCreateSetTheme),
		},
		//todo inline keyboard календарь + время
	}); err != nil {
		return fmt.Errorf("изменение сообщения редактирования задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetDeadline)
}

func (s *Service) ConversationEditTaskSetDeadline(b *gotgbot.Bot, ctx *ext.Context) error {
	deadline, err := time.Parse(types.TimeLayout, ctx.EffectiveMessage.Text)
	if err != nil {
		if _, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("обработка сроков задачи, используйте формат %s", types.TimeLayout), nil); err != nil {
			return fmt.Errorf("ответ про правильный формат сроков: %w", err)
		}
		return handlers.NextConversationState(types.ConversationTaskEditSetDeadline)
	}
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	lastTask.Deadline = deadline
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: 1}}
	themes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем для клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	message := TaskMessageFill("Редактирование задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.CreateThemeInlineKeyboard(lastTask, themes, 1, themesPagesCount),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения редактирования задачи: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения новых сроков задачи")
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetTheme)
}

func (s *Service) ConversationEditTaskSetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор темы")
	}
	callQueryValues := strings.Split(callQuery.Data, ";")
	var themeId, currentPage uint
	for _, callQueryData := range callQueryValues {
		if strings.HasPrefix(callQueryData, types.CallbackTaskThemeChoose) {
			themeStr := strings.Replace(callQueryData, types.CallbackTaskThemeChoose, "", 1)
			theme, err := strconv.ParseUint(themeStr, 10, 64)
			if err != nil {
				return fmt.Errorf("получение идентификатора темы из клавиатуры: %w", err)
			}
			themeId = uint(theme)
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
	themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: currentPage}}
	allThemes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	var chosenTheme types.ThemeModel
	found := false
	for _, theme := range allThemes {
		if theme.ID == themeId {
			chosenTheme = theme
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("выбранная тема не найдена")
	}
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	lastTask.Themes = append(lastTask.Themes, chosenTheme)
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление тем задачи")
	}
	message := TaskMessageFill("Редактирование задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.CreateThemeInlineKeyboard(lastTask, allThemes, currentPage, themesPagesCount),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения редактирования задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetTheme)
}

func (s *Service) ConversationEditTaskDoneTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Выбор завершен"}); err != nil {
		return fmt.Errorf("ответ на завершение выбора: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	message := TaskMessageFill("Редактирование задачи", "Нажмите Создать", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.TaskInlineKeyboard(""),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, завершение выбора тем: %w", err)
	}
	return nil
}

func (s *Service) ConversationEditTaskDone(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Завершение редактирования"}); err != nil {
		return fmt.Errorf("ответ на завершение создания: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней редактируемой задачи: %w", err)
	}
	messageRegister, ok := utils.GetLastMessageRegister(lastTask, types.MessageRegisterOperationEdit)
	if !ok || messageRegister.TaskId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	message := TaskMessageFill("Задача изменена", "", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения редактирования задачи, завершение редактирования: %w", err)
	}
	return handlers.EndConversation()
}
