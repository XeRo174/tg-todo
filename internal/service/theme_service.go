package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"strings"
	"tg-todo/internal/types"
	"tg-todo/internal/utils"
)

// ConversationCreateThemeInit - обработчик разговора инициализации создания темы
func (s *Service) ConversationCreateThemeInit(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := b.SendMessage(ctx.EffectiveSender.ChatId, "Введите имя темы", nil); err != nil {
		return fmt.Errorf("отправка сообщения темы: %w", err)
	}
	return handlers.NextConversationState(types.ConversationNewThemeName)
}

// ConversationCreateThemeSetName - обработчик разговора получения имени темы
func (s *Service) ConversationCreateThemeSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := s.Repository.GetUserByTGId(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск пользователя по tg: %w", err)
	}
	newTheme := types.ThemeModel{
		Name: utils.FirstTitleLetter(ctx.EffectiveMessage.Text),
		User: user,
	}
	if err = s.Repository.CreateTheme(newTheme); err != nil {
		return fmt.Errorf("создание темы: %w", err)
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, "Тема создана", nil); err != nil {
		return fmt.Errorf("отправка сообщения завершения создания: %w", err)
	}
	return handlers.EndConversation()
}

// CommandGetThemes - обработчик команды получения тем
func (s *Service) CommandGetThemes(b *gotgbot.Bot, ctx *ext.Context) error {
	themes, err := s.Repository.GetThemes(types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id})
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
