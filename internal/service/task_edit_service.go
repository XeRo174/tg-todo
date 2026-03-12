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

// ConversationEditTaskSetName - Обработчик разговора установки имени редактируемой задачи
func (s *Service) ConversationEditTaskSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].TaskId == 0 {
		return fmt.Errorf("сообщение редактирования задачи не найдено")
	}
	messageRegister := user.Messages[0]
	task := messageRegister.Task
	task.Name = ctx.EffectiveMessage.Text
	if err = s.Repository.UpdateTask(task); err != nil {
		return fmt.Errorf("обновление имени задачи")
	}
	if err = s.GenerateTaskPriorityMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, "Редактирование задачи", types.ConversationTaskEditSetDeadline); err != nil {
		return err
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаления сообщения нового имени задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetPriority)
}

// ConversationEditTaskSetPriority - обработчик разговора установки приоритета редактируемой задачи
func (s *Service) ConversationEditTaskSetPriority(b *gotgbot.Bot, ctx *ext.Context) error {
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
		return fmt.Errorf("сообщение редактирования задачи не найдено")
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
	if err = s.GenerateTaskDeadlineMessage(b, ctx.EffectiveSender.ChatId, messageRegister, task, task.Deadline, "Редактирование задачи", types.ConversationTaskEditSetTheme); err != nil {
		return err
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetDeadline)
}

// ConversationEditTaskSetTheme - обработчик разговора установки темы редактируемой задачи
func (s *Service) ConversationEditTaskSetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор темы")
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
	task.Themes = append(task.Themes, chosenTheme)
	if err = s.Repository.UpdateTaskThemes(task, task.Themes); err != nil {
		return fmt.Errorf("обновление тем задачи")
	}
	message := TaskMessageFill("Редактирование задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeForTaskInlineKeyboard(task, allThemes, int(themesPagesCount), int(currentPage)),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения редактирования задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetTheme)
}

// ConversationEditTaskUnsetTheme - обработчик разговора удаления темы редактируемой задачи
func (s *Service) ConversationEditTaskUnsetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема отклонена"}); err != nil {
		return fmt.Errorf("ответ на выбор темы")
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
		return fmt.Errorf("обновление тем задачи")
	}
	message := TaskMessageFill("Редактирование задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), task, task.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeForTaskInlineKeyboard(task, allThemes, int(themesPagesCount), int(currentPage)),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения редактирования задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationTaskEditSetTheme)
}
