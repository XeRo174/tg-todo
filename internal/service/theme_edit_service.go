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

// ConversationEditThemeInit - обработчик разговора инициализации редактирования темы
func (s *Service) ConversationEditThemeInit(b *gotgbot.Bot, ctx *ext.Context) error {
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
	themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: 1}}
	themes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем для клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	message := fmt.Sprintf("Выберите тему для редактирования\n\nСтраница: (1/%d)", int(themesPagesCount))
	themeMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeInlineKeyboard(themes, int(themesPagesCount), 1),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения редактирования темы: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(types.MessageRegisterModel{
		UserId:       user.ID,
		BotMessageId: themeMessage.MessageId,
		Operation:    types.MessageRegisterOperationThemeEdit,
	}); err != nil {
		return fmt.Errorf("запись сообщения редактирования темы: %w", err)
	}
	return handlers.NextConversationState(types.ConversationThemeEditChooseTheme)
}

// ConversationEditThemeChoseTheme - обработчик разговора выбора редактируемой темы
func (s *Service) ConversationEditThemeChoseTheme(b *gotgbot.Bot, ctx *ext.Context) error {
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
		return fmt.Errorf("сообщение редактирования темы не найдено")
	}
	themeStr := strings.Replace(callQuery.Data, types.CallbackThemeChoose, "", 1)
	themeId, err := strconv.Atoi(themeStr)
	if err != nil {
		return fmt.Errorf("получение идентификатора темы из клавиатуры: %w", err)
	}
	theme, err := s.Repository.GetThemeById(ctx.EffectiveSender.User.Id, uint(themeId))
	if err != nil {
		return fmt.Errorf("выбранная тема не найдена: %w", err)
	}
	messageRegister := user.Messages[0]
	messageRegister.ThemeId = theme.ID
	message := ThemeMessageFill("Редактирование темы", "Введите новое название", theme)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщение темы, выбор темы для редактирования: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(messageRegister); err != nil {
		return fmt.Errorf("запись сообщения редактирования темы: %w", err)
	}
	return handlers.NextConversationState(types.ConversationThemeEditSetName)
}

// ConversationEditThemeChangeThemesPage - обработчик разговора смены страницы клавиатуры, для выбора редактируемой темы
func (s *Service) ConversationEditThemeChangeThemesPage(b *gotgbot.Bot, ctx *ext.Context) error {
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
		return fmt.Errorf("сообщение редактирования темы не найдено")
	}
	pageStr := strings.Replace(callQuery.Data, types.CallbackChangeThemePage, "", 1)
	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		return fmt.Errorf("получение номера страницы клавиатуры тем: %w", err)
	}
	themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: uint(page)}}
	themes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем для клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	message := fmt.Sprintf("Выберите тему для редактирования.\n\nСтраница: (%d/%d)", page, int(themesPagesCount))
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: user.Messages[0].BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ChooseThemeInlineKeyboard(themes, int(themesPagesCount), int(page)),
		},
	},
	); err != nil {
		return fmt.Errorf("изменение сообщения темы, смена страницы клавиатуры тем: %w", err)
	}
	return nil
}

// ConversationEditThemeSetName - обработчик разговора установки имени темы
func (s *Service) ConversationEditThemeSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].ThemeId == 0 {
		return fmt.Errorf("сообщение редактирования темы не найдено")
	}
	messageRegister := user.Messages[0]
	theme := messageRegister.Theme
	theme.Name = utils.FirstTitleLetter(ctx.EffectiveMessage.Text)
	if err = s.Repository.UpdateTheme(theme); err != nil {
		return fmt.Errorf("обновление имени темы: %w", err)
	}
	message := ThemeMessageFill("Тема обновлена", "", theme)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения темы, завершение редактирования: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения пользователя: %w", err)
	}
	if err = s.Repository.DeleteMessageRegisterByUserId(user.ID); err != nil {
		return fmt.Errorf("очистка сообщенияя редактирования темы: %w", err)
	}
	return handlers.EndConversation()
}
