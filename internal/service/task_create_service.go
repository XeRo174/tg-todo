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

// ConversationCreateTaskInit - обработчик разговора инициализации создания задачи
func (s *Service) ConversationCreateTaskInit(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) > 0 {
		messageRegister := user.Messages[0]
		if _, _, err = b.EditMessageText(MessageOperationBeauty(messageRegister), &gotgbot.EditMessageTextOpts{MessageId: messageRegister.BotMessageId, ChatId: ctx.EffectiveSender.ChatId}); err != nil {
			return fmt.Errorf("изменение прошлого сообщения: %w", err)
		}
	}
	tasks, err := s.Repository.GetTasks(types.TaskFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.UnlimitedSize, Page: 1}})
	if err != nil {
		return fmt.Errorf("получение задач: %w", err)
	}
	newTask := types.TaskModel{
		Name:   fmt.Sprintf("Задача №%d", len(tasks)+1),
		User:   user,
		Status: types.TaskStatusDraft,
	}
	createdTask, err := s.Repository.CreateTask(newTask)
	if err != nil {
		return fmt.Errorf("создание задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", "Введите имя задачи", createdTask, nil)
	taskMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TaskInlineKeyboard(types.ConversationTaskCreateSetPriority),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения задачи: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(types.MessageRegisterModel{
		UserId:       user.ID,
		BotMessageId: taskMessage.MessageId,
		TaskId:       createdTask.ID,
		Operation:    types.MessageRegisterOperationTaskCreate,
	}); err != nil {
		return fmt.Errorf("запись сообщения создания задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskCreateSetName)
}

// ConversationCreateTaskSetName - обработчик разговора получения имени задачи
func (s *Service) ConversationCreateTaskSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение создания задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	task.Name = ctx.EffectiveMessage.Text
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление имени задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", "Выберите приоритет", task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.PriorityButtons(), utils.TaskInlineKeyboard(types.ConversationTaskCreateSetDeadline)...),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, имя задачи: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения имени задачи")
	}
	return handlers.NextConversationState(types.ConversationTaskCreateSetPriority)
}

// ConversationCreateTaskSetPriority - обработчик разговора получения приоритета задачи
func (s *Service) ConversationCreateTaskSetPriority(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Приоритет выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор приоритета: %w", err)
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
	priority, ok := types.ParsePriority(callQuery.Data)
	if !ok {
		priority = types.TaskPriorityNone
	}
	task.Priority = priority
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление приоритета задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Введите срок в формате %s", types.TimeLayout), task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.CreateCalendarButtons(), utils.TaskInlineKeyboard(types.ConversationTaskCreateSetTheme)...),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, приоритет задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskCreateSetDeadline)
}

// ConversationCreateTaskSetDeadline - обработчик разговора получения сроков задачи
func (s *Service) ConversationCreateTaskSetDeadline(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение создания задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	deadline, err := time.Parse(types.TimeLayout, ctx.EffectiveMessage.Text)
	if err != nil {
		if _, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("обработка сроков задачи, используйте формат %s", types.TimeLayout), nil); err != nil {
			return fmt.Errorf("ответ про правильный формат сроков: %w", err)
		}
		return handlers.NextConversationState(types.ConversationTaskCreateSetDeadline)
	}
	task.Deadline = deadline
	if err = s.Repository.UpdateTask(task); err != nil {
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
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeForTaskInlineKeyboard(task, themes, int(themesPagesCount), 1),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, сроки задачи: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения сроков задачи")
	}
	return handlers.NextConversationState(types.ConversationTaskCreateSetTheme)
}

// ConversationCreateTaskSetTheme - обработчик разговора получения тем задачи
func (s *Service) ConversationCreateTaskSetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор темы: %w", err)
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
	callQueryValues := strings.Split(callQuery.Data, ";")
	var themeId, currentPage uint
	for _, callQueryData := range callQueryValues {
		if strings.HasPrefix(callQueryData, types.CallbackTaskSetTheme) {
			themeStr := strings.Replace(callQueryData, types.CallbackTaskSetTheme, "", 1)
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
		return fmt.Errorf("получение тем для клавиатуры: %w", err)
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
	task.Themes = append(task.Themes, chosenTheme)
	if err = s.Repository.UpdateTaskThemes(task, task.Themes); err != nil {
		return fmt.Errorf("обновление тем задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeForTaskInlineKeyboard(task, allThemes, int(themesPagesCount), int(currentPage)),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, темы задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskCreateSetTheme)
}

// ConversationCreateTaskUnsetTheme - обработчик разговора удаления темы создаваемой задачи
func (s *Service) ConversationCreateTaskUnsetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема отклонена"}); err != nil {
		return fmt.Errorf("ответ на выбор темы: %w", err)
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
	callQueryValues := strings.Split(callQuery.Data, ";")
	var themeId, currentPage uint
	for _, callQueryData := range callQueryValues {
		if strings.HasPrefix(callQueryData, types.CallbackTaskUnsetTheme) {
			themeStr := strings.Replace(callQueryData, types.CallbackTaskUnsetTheme, "", 1)
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
		return fmt.Errorf("получение тем для клавиатуры: %w", err)
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
	} else {
		var cleanedThemes []types.ThemeModel
		for _, theme := range task.Themes {
			if theme.ID != chosenTheme.ID {
				cleanedThemes = append(cleanedThemes, theme)
			}
		}
		task.Themes = cleanedThemes
	}
	if err = s.Repository.UpdateTaskThemes(task, task.Themes); err != nil {
		return fmt.Errorf("обновление тем задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeForTaskInlineKeyboard(task, allThemes, int(themesPagesCount), int(currentPage)),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, темы задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskCreateSetTheme)
}
