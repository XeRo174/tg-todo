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

// ThemeMessageFill - формирует сообщение работы с темой
func ThemeMessageFill(title, ending string, theme types.ThemeModel) string {
	return fmt.Sprintf("%s\nНазвание: %s\n\n%s", title, theme.Name, ending)
}

// ConversationThemesInit - обработчик разговора инициализации работы с темой
func (s *Service) ConversationThemesInit(b *gotgbot.Bot, ctx *ext.Context) error {
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
	themeFilter := types.ThemeFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: 1}}
	themes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем для выбора: %w", err)
	}
	themePagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем для выбора: %w", err)
	}
	message := fmt.Sprintf("Ваши темы.\n\nСтраница: (1/%d)", int(themePagesCount))
	themeMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeInlineKeyboard(themes, int(themePagesCount), 1),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения тем: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(types.MessageRegisterModel{
		UserId:       user.ID,
		BotMessageId: themeMessage.MessageId,
		Operation:    types.MessageRegisterOperationTheme,
	}); err != nil {
		return fmt.Errorf("запись сообщения списка тем: %w", err)
	}
	return handlers.NextConversationState(types.ConversationThemeChoose)
}

// ConversationThemeChose - обработчик разговора выбора темы для работы
func (s *Service) ConversationThemeChose(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор темы: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	callQueryValues := strings.Split(callQuery.Data, ";")
	var themeId, currentPage uint
	for _, callQueryData := range callQueryValues {
		if strings.HasPrefix(callQueryData, types.CallbackThemeChoose) {
			themeStr := strings.Replace(callQueryData, types.CallbackThemeChoose, "", 1)
			theme, err := strconv.Atoi(themeStr)
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
	theme, err := s.Repository.GetThemeById(ctx.EffectiveSender.User.Id, themeId)
	if err != nil {
		return fmt.Errorf("выбранная задача не найдена: %w", err)
	}
	messageRegister := user.Messages[0]
	messageRegister.ThemeId = theme.ID
	message := ThemeMessageFill("Информация о теме", "Выберите действие", theme)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ThemeActionButtons(currentPage),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщение темы: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(messageRegister); err != nil {
		return fmt.Errorf("запись сообщения темы: %w", err)
	}
	return handlers.NextConversationState(types.ConversationThemeActionChoose)
}

// ConversationChoseThemeAction - обработчик разговора выбора действия с темой
func (s *Service) ConversationChoseThemeAction(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Действие выбрано"}); err != nil {
		return fmt.Errorf("ответ на выбор действий: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].ThemeId == 0 {
		return fmt.Errorf("сообщение темы не найдено")
	}
	messageRegister := user.Messages[0]
	theme := messageRegister.Theme
	callQueryValues := strings.Split(callQuery.Data, ";")
	var themeAction string
	var currentPage uint
	for _, callQueryData := range callQueryValues {
		if strings.HasPrefix(callQueryData, types.CallbackThemeAction) {
			themeAction = strings.Replace(callQueryData, types.CallbackThemeAction, "", 1)
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
	var message, conversationState string
	var keyboard [][]gotgbot.InlineKeyboardButton
	switch themeAction {
	case types.ActionEdit:
		message = ThemeMessageFill("Редактирование темы", "Введите новое название темы", theme)
		conversationState = types.ConversationThemeEditSetName
		messageRegister.Operation = types.MessageRegisterOperationThemeEdit
	case types.ActionDelete:
		message = ThemeMessageFill("Удаление темы", "Вы точно хотите удалить тему?", theme)
		conversationState = types.ConversationThemeDelete
		messageRegister.Operation = types.MessageRegisterOperationThemeDelete
	case types.ActionBack:
		themeFilter := types.ThemeFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: currentPage}}
		themes, err := s.Repository.GetThemes(themeFilter)
		if err != nil {
			return fmt.Errorf("получение тем для выбора")
		}
		themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
		if err != nil {
			return fmt.Errorf("получение количества страниц тем для выбора: %w", err)
		}
		message = fmt.Sprintf("Ваши темы.\n\nСтраница: (%d/%d)", currentPage, int(themesPagesCount))
		conversationState = types.ConversationThemeChoose
		keyboard = utils.ChooseThemeInlineKeyboard(themes, int(themesPagesCount), int(currentPage))
	default:
		return fmt.Errorf("идентификация типа работы с темой: %s", messageRegister.Operation)
	}
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения темы, редактирование: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(messageRegister); err != nil {
		return fmt.Errorf("запись сообщения '%s': %w", messageRegister.Operation, err)
	}
	return handlers.NextConversationState(conversationState)
}

// CallbackThemeChangeThemesPage - обработчик обратного вызова перехода на следующую страницу клавиатуры тем
func (s *Service) CallbackThemeChangeThemesPage(b *gotgbot.Bot, ctx *ext.Context) error {
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
		return fmt.Errorf("получение номера страницы клавиатуры тем: %w", err)
	}
	themeFilter := types.ThemeFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: uint(page)}}
	themes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем для выбора")
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем для выбора: %w", err)
	}
	message := fmt.Sprintf("Ваши темы.\n\nСтраница: (%d/%d)", page, int(themesPagesCount))
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: user.Messages[0].BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeInlineKeyboard(themes, int(themesPagesCount), page),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения темы, смена страницы клавиатуры тем: %w", err)
	}
	return nil
}
