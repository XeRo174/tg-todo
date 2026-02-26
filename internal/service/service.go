package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/sirupsen/logrus"
	"tg-todo/internal/bootstrap"
	"tg-todo/internal/repository"
	"tg-todo/internal/types"
	"tg-todo/internal/utils"
	"time"
)

type Service struct {
	Repository *repository.Repository
	Bot        *gotgbot.Bot
	App        bootstrap.Application
}

func NewService(r *repository.Repository, bot *gotgbot.Bot, app bootstrap.Application) *Service {
	return &Service{
		Repository: r,
		Bot:        bot,
		App:        app,
	}
}

func (s *Service) Start() {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			logrus.Errorf("Ошибка во время обработки сообщения: %v", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
		Logger:      s.App.Logger,
	})

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{Logger: s.App.Logger})
	dispatcher.AddHandler(handlers.NewCommand("start", s.CommandStartHandler))
	dispatcher.AddHandler(handlers.NewCommand("cancel", s.CommandCancelHandler))
	dispatcher.AddHandler(handlers.NewCommand("get_themes", s.getThemes))
	dispatcher.AddHandler(handlers.NewCommand("get_tasks", s.CommandGetTasks))

	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand("create_theme", s.createTheme)},
		map[string][]ext.Handler{
			utils.ConversationThemeCreateName: {handlers.NewMessage(noCommand, s.setThemeName)},
		}, &handlers.ConversationOpts{}))
	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand("create_task", s.ConversationCreateTaskInit)},
		map[string][]ext.Handler{
			utils.ConversationTaskCreateName:     {handlers.NewMessage(noCommand, s.ConversationCreateTaskSetName)},
			utils.ConversationTaskCreatePriority: {handlers.NewCallback(callbackquery.Prefix("task_priority:"), s.ConversationCreateTaskSetPriority)},
			utils.ConversationTaskCreateDeadline: {handlers.NewMessage(noCommand, s.ConversationCreateTaskSetDeadline)},
			utils.ConversationTaskCreateTheme:    {handlers.NewMessage(noCommand, s.ConversationCreateTaskSetTheme)},
		},
		&handlers.ConversationOpts{}))

	err := updater.StartPolling(s.Bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		s.App.Logger.Error("Ошибка получения сообщения: ", err)
	}
	s.App.Logger.Info("Bot has been started...", "bot_username", s.Bot.User.Username)

	updater.Idle()
}

func (s *Service) CommandStartHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	newUser := types.UserModel{
		TGId:     ctx.EffectiveSender.User.Id,
		Username: ctx.EffectiveSender.User.Username,
	}
	if err := s.Repository.CreateUser(newUser); err != nil {
		return fmt.Errorf("ошибка создания нового пользователя: %v", err)
	}
	//if err := s.Repository.Create(&newUser).Error; err != nil {
	//	return fmt.Errorf("ошибка создания нового пользователя: %v", err)
	//}
	if _, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Пользователь создан"), &gotgbot.SendMessageOpts{}); err != nil {
		return fmt.Errorf("ошибка отправки стартового сообщения: %v", err)
	}
	return nil
}

func (s *Service) CommandCancelHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	return handlers.EndConversation()
}

func noCommand(msg *gotgbot.Message) bool {
	return message.Text(msg) && !message.Command(msg)
}
