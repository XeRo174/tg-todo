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
		return fmt.Errorf("ошибка отправки сообщения получения имени: %v", err)
	}
	return handlers.NextConversationState(utils.ConversationThemeCreateName)
}

// ConversationCreateThemeSetName - обработчик разговора получения имени темы
func (s *Service) ConversationCreateThemeSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := s.Repository.GetUserByTGId(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("ошибка получения пользователя: %v", err)
	}
	newTheme := types.ThemeModel{
		Name: utils.FirstTitleLetter(ctx.EffectiveMessage.Text),
		User: user,
	}
	if err = s.Repository.CreateTheme(newTheme); err != nil {
		return fmt.Errorf("ошибка создания темы: %v", err)
	}
	b.SendMessage(ctx.EffectiveSender.ChatId, "Тема создана", nil)
	return handlers.EndConversation()
}

// CommandGetThemes - обработчик команды получения тем
func (s *Service) CommandGetThemes(b *gotgbot.Bot, ctx *ext.Context) error {
	themes, err := s.Repository.GetThemes(types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id})
	if err != nil {
		return fmt.Errorf("ошибка получения тем: %v", err)
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
		return fmt.Errorf("ошибка отправки тем: %v", err)
	}
	return nil
}
