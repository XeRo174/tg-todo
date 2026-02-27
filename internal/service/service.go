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

//todo - перевести выбор тем на кнопки под сообщением с постраничным выбором. Реализация - заканчивается этап установки сроков задачи, получаю список первых N тем, формирую строчную клавиатуру с каждой темой как клавишу вида []Покупки,[]Учеба,[]Прочее.
// Последняя строка набор клавиш - переключателей страниц вида <-,->. Нажатие на переключателей меняет кнопки Тем на следующие. Если нет прошлой или следующей страницы тем, то соответствующей кнопки нет.
// Нажатие на кнопку темы редактирует текст сообщения добавляя туда выбранную тему и клавиша темы отмечается [x].

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

//todo - добавить уникальность Тем исходя из Имени и id пользователя.

func (s *Service) Start() {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			logrus.Errorf("Ошибка во время обработки сообщения: %v", err)
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

	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand("create_theme", s.ConversationCreateThemeInit)},
		map[string][]ext.Handler{
			utils.ConversationThemeCreateName: {handlers.NewMessage(noCommand, s.ConversationCreateThemeSetName)},
		}, &handlers.ConversationOpts{
			Exits:        []ext.Handler{cancelCommand},
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		}))
	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand("create_task", s.ConversationCreateTaskInit)},
		map[string][]ext.Handler{
			utils.ConversationTaskCreateName:     {handlers.NewMessage(noCommand, s.ConversationCreateTaskSetName)},
			utils.ConversationTaskCreatePriority: {handlers.NewCallback(callbackquery.Prefix("task_priority:"), s.ConversationCreateTaskSetPriority)},
			utils.ConversationTaskCreateDeadline: {handlers.NewMessage(noCommand, s.ConversationCreateTaskSetDeadline)},
			utils.ConversationTaskCreateTheme:    {handlers.NewMessage(noCommand, s.ConversationCreateTaskSetTheme)},
		},
		&handlers.ConversationOpts{
			Exits:        []ext.Handler{cancelCommand},
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
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
		return fmt.Errorf("ошибка создания нового пользователя: %v", err)
	}
	if _, err := b.SendMessage(ctx.EffectiveSender.ChatId, "Пользователь создан", nil); err != nil {
		return fmt.Errorf("ошибка отправки стартового сообщения: %v", err)
	}
	return nil
}

// CommandCancelHandler - обработчик отмены действий/разговора
func (s *Service) CommandCancelHandler(b *gotgbot.Bot, ctx *ext.Context) error {
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
		Name: "Прочее",
	}
	s.Repository.CreateTheme(theme1)
	s.Repository.CreateTheme(theme2)
	s.Repository.CreateTheme(theme3)
	return nil
}

func noCommand(msg *gotgbot.Message) bool {
	return message.Text(msg) && !message.Command(msg)
}
