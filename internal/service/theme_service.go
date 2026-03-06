package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"strings"
	"tg-todo/internal/types"
)

// ThemeMessageFill - формирует сообщение работы с темой
func ThemeMessageFill(title, ending string, theme types.ThemeModel) string {
	return fmt.Sprintf("%s\nНазвание: %s\n\n%s", title, theme.Name, ending)
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
