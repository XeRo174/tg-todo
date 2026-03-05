package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"tg-todo/internal/types"
	"tg-todo/internal/utils"
)

// ConversationCreateThemeInit - обработчик разговора инициализации создания темы
func (s *Service) ConversationCreateThemeInit(b *gotgbot.Bot, ctx *ext.Context) error {
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
	themes, err := s.Repository.GetThemes(types.ThemeFilter{UserTGId: userTGId, SortQuery: types.SortQuery{Size: types.ThemeKeyboardSize, Page: 1}})
	if err != nil {
		return fmt.Errorf("получение тем для клавиатуры выбора: %w", err)
	}
	newTheme := types.ThemeModel{
		Name: fmt.Sprintf("Тема №%d", len(themes)+1),
		User: user,
	}
	createdTheme, err := s.Repository.CreateTheme(newTheme)
	if err != nil {
		return fmt.Errorf("создание темы: %w", err)
	}
	message := ThemeMessageFill("Создание темы", "Введите имя темы", createdTheme)
	themeMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, nil)
	if err != nil {
		return fmt.Errorf("отправка сообщения темы: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(types.MessageRegisterModel{
		UserId:       user.ID,
		BotMessageId: themeMessage.MessageId,
		ThemeId:      createdTheme.ID,
		Operation:    types.MessageRegisterOperationThemeCreate,
	}); err != nil {
		return fmt.Errorf("запись сообщения создания темы: %w", err)
	}
	return handlers.NextConversationState(types.ConversationThemeCreateSetName)
}

// ConversationCreateThemeSetName - обработчик разговора получения имени темы
func (s *Service) ConversationCreateThemeSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 || user.Messages[0].ThemeId == 0 {
		return fmt.Errorf("сообщение создания темы не найдено")
	}
	messageRegister := user.Messages[0]
	theme := messageRegister.Theme
	theme.Name = utils.FirstTitleLetter(ctx.EffectiveMessage.Text)
	if err = s.Repository.UpdateTheme(theme); err != nil {
		return fmt.Errorf("изменение имени темы")
	}
	message := ThemeMessageFill("Тема создана", "", theme)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения темы, завершение создания: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения имени темы")
	}
	if err = s.Repository.DeleteMessageRegisterByUserTGId(userTGId); err != nil {
		return fmt.Errorf("очистка сообщения создания темы: %w", err)
	}
	return handlers.EndConversation()
}
