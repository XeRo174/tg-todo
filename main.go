package main

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"tg-todo/internal/bootstrap"
	"tg-todo/internal/repository"
	"tg-todo/internal/service"
)

func main() {

	app := bootstrap.App()
	r := repository.NewRepository(app.Database)
	bot, err := gotgbot.NewBot(app.Environment.Token, nil)
	if err != nil {
		app.Logger.Error(fmt.Sprintf("Ошибка создания бота, err: %v", err))
		return
	}
	s := service.NewService(r, bot, app)

	s.Start()
	//
	//
	//dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
	//	Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
	//		logrus.Errorf("Ошибка во время обработки сообщения: %v", err)
	//		return ext.DispatcherActionNoop
	//	},
	//	MaxRoutines: ext.DefaultMaxRoutines,
	//	Logger:      logger,
	//})
	//updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{Logger: logger})
	//dispatcher.AddHandler(handlers.NewCommand("start", service.start))
	//dispatcher.AddHandler(handlers.NewCommand("cancel", service.cancel))
	//dispatcher.AddHandler(handlers.NewCommand("get_themes", service.getThemes))
	//dispatcher.AddHandler(handlers.NewCommand("get_tasks", service.getTasks))
	//
	//dispatcher.AddHandler(handlers.NewConversation(
	//	[]ext.Handler{handlers.NewCommand("create_theme", service.createTheme)},
	//	map[string][]ext.Handler{
	//		ConversationThemeCreateName: {handlers.NewMessage(noCommand, service.setThemeName)},
	//	}, &handlers.ConversationOpts{}))
	//dispatcher.AddHandler(handlers.NewConversation(
	//	[]ext.Handler{handlers.NewCommand("create_task", service.createTask)},
	//	map[string][]ext.Handler{
	//		ConversationTaskCreateName:     {handlers.NewMessage(noCommand, service.setTaskName)},
	//		ConversationTaskCreatePriority: {handlers.NewCallback(callbackquery.Prefix("task_priority:"), service.setTaskPriority)},
	//		ConversationTaskCreateDeadline: {handlers.NewMessage(noCommand, service.setTaskDeadline)},
	//		ConversationTaskCreateTheme:    {handlers.NewMessage(noCommand, service.setTaskTheme)},
	//	},
	//	&handlers.ConversationOpts{}))
	//
	//err = updater.StartPolling(bot, &ext.PollingOpts{
	//	DropPendingUpdates: true,
	//	GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
	//		Timeout: 9,
	//		RequestOpts: &gotgbot.RequestOpts{
	//			Timeout: time.Second * 10,
	//		},
	//	},
	//})
	//if err != nil {
	//	logger.Error("Ошибка получения сообщения: ", err)
	//}
	//logger.Info("Bot has been started...", "bot_username", bot.User.Username)
	//
	//updater.Idle()
}
