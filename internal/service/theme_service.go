package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
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

// Новый формат функций

func (s *Service) ChooseThemeInlineKeyboard(themesByPage []types.ThemeModel, totalPages, currentPage int) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	for _, theme := range themesByPage {
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: theme.Name, CallbackData: fmt.Sprintf("%s%d", types.CallbackThemeChoose, theme.ID)},
		})
	}
	if totalPages > 1 {
		buttons = append(buttons, utils.CreateArrowButtons(currentPage, totalPages)...)
	}
	return buttons
}
