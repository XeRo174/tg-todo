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

func (s *Service) CommandUserTimezoneHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	timezoneMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, "Установка часовой зоны пользователя", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TimezonesInlineKeyboard(),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения часовых зон: %w", err)
	}
	if err = s.Repository.UpsertMessageRegister(types.MessageRegisterModel{
		UserId:       user.ID,
		BotMessageId: timezoneMessage.MessageId,
		Operation:    types.MessageRegisterOperationUserEdit,
	}); err != nil {
		return fmt.Errorf("запись сообщения редактирования пользователя: %w", err)
	}
	return handlers.NextConversationState(types.ConversationUserEditChooseTimezone)
}

func (s *Service) CallbackUserTimezoneChoose(b *gotgbot.Bot, ctx *ext.Context) error {
	userTGId := ctx.EffectiveSender.User.Id
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Часовая зона выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор часовой зоны: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(userTGId)
	if err != nil {
		return fmt.Errorf("%s: %w", types.ErrorStrokeFindUserByTG, err)
	}
	if len(user.Messages) == 0 || user.Messages[0].BotMessageId == 0 {
		return fmt.Errorf("сообщение создания задачи не найдено")
	}
	messageRegister := user.Messages[0]
	timezoneStr := strings.Replace(callQuery.Data, types.CallbackUserSetTimezone, "", 1)
	user.TimeZone = timezoneStr
	if err = s.Repository.UpdateUser(user); err != nil {
		return fmt.Errorf("обновление часовой зоны пользователя: %w", err)
	}
	if _, _, err = b.EditMessageText(fmt.Sprintf("Установка часовой зоны пользователя завершена, выбрана зона '%s'", timezoneStr), &gotgbot.EditMessageTextOpts{
		MessageId: messageRegister.BotMessageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения пользователя, установка часовой зоны завершена: %w", err)
	}
	return handlers.EndConversation()
}
