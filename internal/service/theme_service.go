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

// createTheme - обработчик инициализации создания темы
func (s *Service) createTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, "Введите имя темы", &gotgbot.SendMessageOpts{})
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения получения имени: %v", err)
	}
	return handlers.NextConversationState(utils.ConversationThemeCreateName)
}

// setThemeName - обработчик получения имени темы и записи в бд
func (s *Service) setThemeName(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := s.GetUserByTGId(ctx.EffectiveSender.User.Id)
	if err != nil {
		return err
	}
	newTheme := types.TaskThemeModel{
		Name: ctx.EffectiveMessage.Text,
		User: user,
	}
	if err := s.Database.Create(&newTheme).Error; err != nil {
		return fmt.Errorf("ошибка создания темы: %v", err)
	}
	if _, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Тема создана"), &gotgbot.SendMessageOpts{}); err != nil {
		return fmt.Errorf("ошибка отправки сообщения создания темы: %v", err)
	}
	return handlers.EndConversation()
}

func (s *Service) getThemes(b *gotgbot.Bot, ctx *ext.Context) error {
	var themes []types.TaskThemeModel
	if err := s.Database.Joins("join user_models ON user_models.id = task_theme_models.user_id").Where("user_models.tg_id=?", ctx.EffectiveSender.User.Id).Find(&themes).Error; err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf("ошибка получения тем: %v", err), &gotgbot.SendMessageOpts{})
		return handlers.NextConversationState(ConversationTaskCreateTheme)
	}
	var themeStroke []string
	for _, theme := range themes {
		themeStroke = append(themeStroke, fmt.Sprintf("Тема: %s", theme.Name))
	}
	if _, err := ctx.EffectiveMessage.Reply(b, strings.Join(themeStroke, "\n"), &gotgbot.SendMessageOpts{}); err != nil {
		return fmt.Errorf("getThemes error:  %v", err)
	}
	return nil
}
