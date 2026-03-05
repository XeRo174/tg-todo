package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/sirupsen/logrus"
	"tg-todo/internal/bootstrap"
	"tg-todo/internal/repository"
	"tg-todo/internal/types"
	"time"
)

// ThemeSession - структура для хранения сессии редактирования или создания темы
type ThemeSession struct {
	ThemeId   uint
	MessageId int64
}

type Service struct {
	Repository        *repository.Repository
	Bot               *gotgbot.Bot
	App               bootstrap.Application
	ThemeEditSessions map[int64]ThemeSession
}

func NewService(r *repository.Repository, bot *gotgbot.Bot, app bootstrap.Application) *Service {
	return &Service{
		Repository:        r,
		Bot:               bot,
		App:               app,
		ThemeEditSessions: make(map[int64]ThemeSession),
	}
}

//todo - поиск удобной версии установки сроков задача. Текущая вариант не удобен и только выполняет свое предназначение. Нужен виджет/команда/кнопки где будет более подходящий способ установки сроков задач.

//todo - возможность пропускать ненужные параметры для задачи.

//todo - редактирование тем и задач.

//todo - уменьшение количества отдельных сообщений. При создании задачи, отправляется слишком много отдельных сообщений-вопросов. Решение - в идеале, после получения /create_task отправляется одно сообщение, которое будет последовательно редактироваться дополняясь новыми данными.
// Примерный вид "Создание новой задачи\nИмя задачи: Купить учебник по sql\nСроки выполнения: 04.03.2026\n\n\nВведите приоритет задачи"
// Под сообщением будет ряд кнопок: Создать - завершить создание с текущими данными; Редактировать - появляются кнопки каждого поля, после нажатия надо ввести-выбрать новое значение и процесс продолжается; Отменить - черновик задачи удалятся из базы, процесс завершен. В зависимости будут показываться и скрываться разные кнопки.

//todo - отметка статуса задачи. Пользователь может отметить задачу выполненной, может установить её как заброшенную и тому подобные статусы. По умолчанию будет ставится статус в работе или подобный.

//todo - менеджер уведомления задач. Для уведомления пользователя о сроках задачи нужен Менеджер работающий в отдельном потоке (goroutine). Он будет периодически обращаться к бд и получать все не выполненные задачи.
// Также нужно поле у Задачи, отвечающее за количество уведомлений, в нем будет считаться за какое время было уведомление, за день, за час, за 10 минут. (Условный пример). Если остался час до окончания задачи и подобное сообщение не было прежде отправлено, то Менеджер подает отправляет сообщение пользователю.

//todo - Получение часовой зоны пользователя. Чтобы уведомления о сроках задачи отправлялись в правильное время нужно учитывать время пользователя.
// Иначе, если пользователь живет по GMT+5 и ставит сроки задачи 14:00:00, а Сервер будет работать в Москве, то он отправит уведомление "Остался час" в 13:00:00 по Москве, а у пользователя будет уже 15:00:00.
// Для получения этих данных можно спросить у пользователя его текущее время в 24 часом варианте, так получим возможность сравнить его с gmt и установить в профиле пользователя.

func (s *Service) Start() {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			//todo отправка сообщения пользователю
			logrus.Errorf("Ошибка во время обработки сообщения: %+v", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
		Logger:      s.App.Logger,
	})
	cancelCommand := handlers.NewCommand("cancel", s.CommandCancelHandler)
	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{Logger: s.App.Logger})
	dispatcher.AddHandler(handlers.NewCommand("start", s.CommandStartHandler))
	dispatcher.AddHandler(handlers.NewCommand("common", s.CommandCommonValue))
	dispatcher.AddHandler(cancelCommand)
	dispatcher.AddHandler(handlers.NewCommand("get_themes", s.CommandGetThemes))
	dispatcher.AddHandler(handlers.NewCommand("get_tasks", s.CommandGetTasks))

	//Разговор создания новой темы
	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand(types.CommandThemeCreateInit, s.ConversationCreateThemeInit)},
		map[string][]ext.Handler{
			types.ConversationThemeCreateSetName: {handlers.NewMessage(noCommand, s.ConversationCreateThemeSetName)},
		}, &handlers.ConversationOpts{
			Exits:        []ext.Handler{cancelCommand},
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		}))
	//Разговор редактирования темы
	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand(types.CommandThemeEditInit, s.ConversationEditThemeInit)},
		map[string][]ext.Handler{
			types.ConversationThemeEditChooseTheme: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackThemeEditChooseTheme), s.ConversationEditThemeChoseTheme),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackThemeEditChangeThemesPage), s.ConversationEditThemeChangeThemesPage),
			},
			types.ConversationThemeEditSetName: {handlers.NewMessage(noCommand, s.ConversationEditThemeSetName)},
		},
		&handlers.ConversationOpts{
			Exits:        []ext.Handler{cancelCommand},
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		},
	))

	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand(types.CommandTaskCreateInit, s.ConversationCreateTaskInit)},
		map[string][]ext.Handler{
			types.ConversationTaskCreateSetName: {
				handlers.NewMessage(noCommand, s.ConversationCreateTaskSetName),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.ConversationCreateTaskSkipField),
			},
			types.ConversationTaskCreateSetPriority: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskPrioritySet), s.ConversationCreateTaskSetPriority),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.ConversationCreateTaskSkipField),
			},
			types.ConversationTaskCreateSetDeadline: {
				handlers.NewMessage(noCommand, s.ConversationCreateTaskSetDeadline),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.ConversationCreateTaskSkipField),
			},
			types.ConversationTaskCreateSetTheme: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskThemeChoose), s.ConversationCreateTaskSetTheme),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackChangeTaskThemesPage), s.ConversationCreateTaskChangePage),
				handlers.NewCallback(callbackquery.Equal(types.CallbackThemeChoseDone), s.ConversationCreateTaskDoneTheme),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.ConversationCreateTaskSkipField),
			},
		},
		&handlers.ConversationOpts{
			Exits: []ext.Handler{
				cancelCommand,
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskCreateDone), s.ConversationCreateTaskDone),
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskCreateCancel), s.ConversationCreateTaskCancel),
			},
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		}))

	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand(types.CommandTaskEditInit, s.ConversationEditTaskInit)},
		map[string][]ext.Handler{
			types.ConversationTaskEditChooseTask: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskChoose), s.ConversationEditTaskChooseTask),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackChangePage), s.CallbackTaskChangeTasksPage),
			},
			types.ConversationTaskEditSetName: {
				handlers.NewMessage(noCommand, s.ConversationEditTaskSetName),
			},
			types.ConversationTaskEditSetPriority: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskPrioritySet), s.ConversationEditTaskSetName),
			},
			types.ConversationTaskEditSetDeadline: {
				handlers.NewMessage(noCommand, s.ConversationEditTaskSetDeadline),
			},
			types.ConversationTaskEditSetTheme: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskThemeChoose), s.ConversationEditTaskSetTheme),
			},
		},
		&handlers.ConversationOpts{
			Exits: []ext.Handler{
				cancelCommand,
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskDone), s.CallbackTaskDone),
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskCancel), s.CallbackTaskCancel),
			},
		}))

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

// CommandStartHandler - обработчик команды запуска бота пользователем
func (s *Service) CommandStartHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	newUser := types.UserModel{
		TGId:     ctx.EffectiveSender.User.Id,
		Username: ctx.EffectiveSender.User.Username,
	}
	if err := s.Repository.CreateUser(newUser); err != nil {
		return fmt.Errorf("создание нового пользователя: %w", err)
	}
	if _, err := b.SendMessage(ctx.EffectiveSender.ChatId, "Пользователь создан", nil); err != nil {
		return fmt.Errorf("отправка стартового сообщения: %w", err)
	}
	if err := s.CommandCommonValue(b, ctx); err != nil {
		return fmt.Errorf("установка базовых тем: %w", err)
	}
	return nil
}

// CommandCancelHandler - обработчик отмены действий/разговора
func (s *Service) CommandCancelHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := b.SendMessage(ctx.EffectiveSender.ChatId, "Разговор завершен", nil); err != nil {
		return fmt.Errorf("отправка сообщения отмены: %w", err)
	}
	return handlers.EndConversation()
}

// CommandCommonValue - обработчик команды записи базовых тем
func (s *Service) CommandCommonValue(b *gotgbot.Bot, ctx *ext.Context) error {
	user, _ := s.Repository.GetUserByTGId(ctx.EffectiveSender.User.Id)

	theme1 := types.ThemeModel{
		User: user,
		Name: "Покупки",
	}
	theme2 := types.ThemeModel{
		User: user,
		Name: "Учеба",
	}
	theme3 := types.ThemeModel{
		User: user,
		Name: "Работа",
	}
	theme4 := types.ThemeModel{
		User: user,
		Name: "Дом",
	}
	theme5 := types.ThemeModel{
		User: user,
		Name: "Прочее",
	}
	if _, err := s.Repository.CreateTheme(theme1); err != nil {
		return fmt.Errorf("создание темы 1: %w", err)
	}
	if _, err := s.Repository.CreateTheme(theme2); err != nil {
		return fmt.Errorf("создание темы 2: %w", err)
	}
	if _, err := s.Repository.CreateTheme(theme3); err != nil {
		return fmt.Errorf("создание темы 3: %w", err)
	}
	if _, err := s.Repository.CreateTheme(theme4); err != nil {
		return fmt.Errorf("создание темы 4: %w", err)
	}
	if _, err := s.Repository.CreateTheme(theme5); err != nil {
		return fmt.Errorf("создание темы 5: %w", err)
	}
	return nil
}

func noCommand(msg *gotgbot.Message) bool {
	return message.Text(msg) && !message.Command(msg)
}

func MessageOperationBeauty(messageRegister types.MessageRegisterModel) string {
	var title string
	switch messageRegister.Operation {
	case types.MessageRegisterOperationTaskCreate:
		title = "Создание задачи прервано"
	case types.MessageRegisterOperationTaskEdit:
		title = "Редактирование задачи прервано"
	case types.MessageRegisterOperationThemeCreate:
		title = "Создание темы прервано"
	case types.MessageRegisterOperationThemeEdit:
		title = "Редактирование темы прервано"
	default:
		title = "Работа прервана"
	}
	if messageRegister.TaskId != 0 {
		return TaskMessageFill(title, "", messageRegister.Task, messageRegister.Task.Themes)
	} else if messageRegister.ThemeId != 0 {
		return ThemeMessageFill(title, "", messageRegister.Theme)
	} else {
		return title
	}

}
