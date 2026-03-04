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

func ThemeMessageFill(title, ending string, theme types.ThemeModel) string {
	return fmt.Sprintf("%s\nНазвание: %s\n\n%s", title, theme.Name, ending)
}

func (s *Service) ClearThemeSession(userTGId int64) {
	delete(s.ThemeEditSessions, userTGId)
}

// ConversationCreateThemeInit - обработчик разговора инициализации создания темы
func (s *Service) ConversationCreateThemeInit(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	if session, ok := s.ThemeEditSessions[userTGId]; ok && session.MessageId != 0 {
		theme, err := s.Repository.GetThemeById(userTGId, session.ThemeId)
		if err != nil {
			return fmt.Errorf("поиск последней темы")
		}
		message := ThemeMessageFill("Тема", "отмена прошлых действий", theme)
		if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
			MessageId: session.MessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
		}); err != nil {
			return fmt.Errorf("изменение сообщения темы: %w", err)
		}
	}

	themes, err := s.Repository.GetThemes(types.ThemeFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.UnlimitedSize, Page: 1}})
	if err != nil {
		return fmt.Errorf("получение тем: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("поиск пользователя по tg: %w", err)
	}
	newTheme := types.ThemeModel{
		Name: fmt.Sprintf("Тема №%d", len(themes)+1),
		User: user,
	}
	createdTheme, err := s.Repository.CreateTheme(newTheme)
	if err != nil {
		return fmt.Errorf("создание темы: %w", err)
	}
	message := ThemeMessageFill("Создание темы", "Введите имя темы", newTheme)
	themeMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, nil)
	if err != nil {
		return fmt.Errorf("отправка сообщения темы: %w", err)
	}
	s.ThemeEditSessions[userTGId] = ThemeSession{
		ThemeId:   createdTheme.ID,
		MessageId: themeMessage.MessageId,
	}
	return handlers.NextConversationState(types.ConversationThemeCreateSetName)
}

// ConversationCreateThemeSetName - обработчик разговора получения имени темы
func (s *Service) ConversationCreateThemeSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	themeSession, ok := s.ThemeEditSessions[userTGId]
	if !ok || themeSession.ThemeId == 0 {
		return fmt.Errorf("сессия создания темы не найдена")
	}
	theme, err := s.Repository.GetThemeById(userTGId, themeSession.ThemeId)
	if err != nil {
		return fmt.Errorf("поиск темы: %w", err)
	}
	theme.Name = utils.FirstTitleLetter(ctx.EffectiveMessage.Text)
	if err = s.Repository.UpdateTheme(theme); err != nil {
		return fmt.Errorf("изменение имени темы")
	}
	message := ThemeMessageFill("Тема создана", "", theme)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: themeSession.MessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения темы: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения имени темы")
	}
	s.ClearThemeSession(userTGId)
	return handlers.EndConversation()
}

// CommandGetThemes - обработчик команды получения тем
func (s *Service) CommandGetThemes(b *gotgbot.Bot, ctx *ext.Context) error {
	themes, err := s.Repository.GetThemes(types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: types.UnlimitedSize, Page: 1}})
	if err != nil {
		return fmt.Errorf("получение тем: %w", err)
	}
	var themeStroke []string
	for _, theme := range themes {
		themeStroke = append(themeStroke, fmt.Sprintf("Тема: %s", theme.Name))
	}
	var message string
	if len(themeStroke) > 0 {
		message = strings.Join(themeStroke, "\n")
	} else {
		message = "Нет тем"
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, message, nil); err != nil {
		return fmt.Errorf("отправка тем: %w", err)
	}
	return nil
}

func (s *Service) ConversationEditThemeInit(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	if session, ok := s.ThemeEditSessions[userTGId]; ok && session.MessageId != 0 {
		theme, err := s.Repository.GetThemeById(userTGId, session.ThemeId)
		if err != nil {
			return fmt.Errorf("поиск последней темы")
		}
		message := ThemeMessageFill("Тема", "отмена прошлых действий", theme)
		if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
			MessageId: session.MessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
		}); err != nil {
			return fmt.Errorf("изменение сообщения темы: %w", err)
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
			InlineKeyboard: s.CreateThemePagesInlineKeyboard(themes, 1, themesPagesCount),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения редактирования темы: %w", err)
	}
	s.ThemeEditSessions[userTGId] = ThemeSession{
		ThemeId:   0,
		MessageId: themeMessage.MessageId,
	}
	return handlers.NextConversationState(types.ConversationThemeEditChooseTheme)
}

func (s *Service) ConversationEditThemeChoseTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор темы: %w", err)
	}
	themeSession, ok := s.ThemeEditSessions[userTGId]
	if !ok {
		return fmt.Errorf("сессия редактирования темы не найдена")
	}
	themeStr := strings.Replace(callQuery.Data, types.CallbackThemeEditChooseTheme, "", 1)
	themeId, err := strconv.Atoi(themeStr)
	if err != nil {
		return fmt.Errorf("получение идентификатора темы из клавиатуры: %w", err)
	}
	theme, err := s.Repository.GetThemeById(ctx.EffectiveSender.User.Id, uint(themeId))
	if err != nil {
		return fmt.Errorf("выбранная тема не найдена: %w", err)
	}
	s.ThemeEditSessions[userTGId] = ThemeSession{
		ThemeId:   theme.ID,
		MessageId: themeSession.MessageId,
	}
	message := ThemeMessageFill("Редактирование темы", "Введите новое название", theme)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: themeSession.MessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщение темы: %w", err)
	}
	return handlers.NextConversationState(types.ConversationThemeEditSetName)
}

func (s *Service) ConversationEditThemeChangeThemesPage(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Страница тем изменена"}); err != nil {
		return fmt.Errorf("ответ на переключение страницы: %w", err)
	}
	themeSession, ok := s.ThemeEditSessions[userTGId]
	if !ok {
		return fmt.Errorf("сессия редактирования темы не найдена")
	}
	pageStr := strings.Replace(callQuery.Data, types.CallbackThemeEditChangeThemesPage, "", 1)
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
	if _, _, err = b.EditMessageText(
		message,
		&gotgbot.EditMessageTextOpts{
			MessageId: themeSession.MessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: s.CreateThemePagesInlineKeyboard(themes, uint(page), themesPagesCount),
			},
		},
	); err != nil {
		return fmt.Errorf("изменение страницы клавиатуры тем: %w", err)
	}
	return nil
}

func (s *Service) ConversationEditThemeSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	session, ok := s.ThemeEditSessions[userTGId]
	if !ok || session.ThemeId == 0 {
		return fmt.Errorf("сессия редактирования темы не найдена")
	}
	theme, err := s.Repository.GetThemeById(userTGId, session.ThemeId)
	if err != nil {
		return fmt.Errorf("поиск темы: %w", err)
	}
	theme.Name = utils.FirstTitleLetter(ctx.EffectiveMessage.Text)
	if err = s.Repository.UpdateTheme(theme); err != nil {
		return fmt.Errorf("обновление имени темы: %w", err)
	}
	message := ThemeMessageFill("Тема обновлена", "", theme)
	if _, _, err = b.EditMessageText(
		message,
		&gotgbot.EditMessageTextOpts{
			MessageId: session.MessageId,
			ChatId:    ctx.EffectiveSender.ChatId,
		},
	); err != nil {
		return fmt.Errorf("изменение сообщения темы: %w", err)
	}

	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения пользователя: %w", err)
	}

	s.ClearThemeSession(userTGId)
	return handlers.EndConversation()
}

// CreateThemePagesInlineKeyboard - создает простую клавиатуру выбора темы для редактирования
func (s *Service) CreateThemePagesInlineKeyboard(themesByPage []types.ThemeModel, page uint, pagesCount float64) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, theme := range themesByPage {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: theme.Name, CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChooseTheme, theme.ID)},
		})
	}
	totalPages := int(pagesCount)
	currentPage := int(page)
	if totalPages > 1 {
		if currentPage == 1 {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", totalPages, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChangeThemesPage, totalPages)},
				{Text: fmt.Sprintf("(%d/%d) >", currentPage+1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChangeThemesPage, currentPage+1)},
			})
		} else if currentPage == totalPages {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChangeThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", 1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChangeThemesPage, 1)},
			})
		} else if currentPage == totalPages {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChangeThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", 1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChangeThemesPage, 1)},
			})
		} else {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChangeThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", currentPage+1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeEditChangeThemesPage, currentPage+1)},
			})
		}
	}
	return buttons
}
