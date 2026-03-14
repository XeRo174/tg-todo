package service

import (
	"errors"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"tg-todo/internal/bootstrap"
	"tg-todo/internal/repository"
	"tg-todo/internal/types"
	"time"
)

type Service struct {
	Repository    *repository.Repository
	Bot           *gotgbot.Bot
	App           bootstrap.Application
	SignalChan    chan struct{}
	CheckInterval time.Duration
}

func NewService(r *repository.Repository, bot *gotgbot.Bot, app bootstrap.Application, checkInterval time.Duration) *Service {
	return &Service{
		Repository:    r,
		Bot:           bot,
		App:           app,
		CheckInterval: checkInterval,
		SignalChan:    make(chan struct{}),
	}
}

//todo - менеджер уведомления задач. Для уведомления пользователя о сроках задачи нужен Менеджер работающий в отдельном потоке (goroutine). Он будет периодически обращаться к бд и получать все не выполненные задачи.
// Также нужно поле у Задачи, отвечающее за количество уведомлений, в нем будет считаться за какое время было уведомление, за день, за час, за 10 минут. (Условный пример). Если остался час до окончания задачи и подобное сообщение не было прежде отправлено, то Менеджер подает отправляет сообщение пользователю.

func (s *Service) Start() {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			var convErr *handlers.ConversationStateChange
			if errors.As(err, &convErr) && convErr.End {
				return ext.DispatcherActionNoop
			}
			//errorMessage := fmt.Sprintf("Ошибка по время обработки сообщения: %+v", err)
			//adminMessage, err := b.SendMessage(adminTGId, errorMessage, nil)
			if _, err := b.SendMessage(ctx.EffectiveSender.ChatId, "Во время обработки произошла ошибка", nil); err != nil {
				s.App.Logger.Warn(fmt.Sprintf("Ошибка отправки пользователю сообщение предупреждение: %v", err))
			}
			s.App.Logger.Error(fmt.Sprintf("ошибка во время обработки сообщения: %+v", err))

			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
		Logger:      s.App.Logger,
	})

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{Logger: s.App.Logger})
	startCommand := handlers.NewCommand(types.CommandStart, s.CommandStartHandler)
	//commonCommand := handlers.NewCommand("common", s.CommandCommonValue)
	cancelCommand := handlers.NewCommand(types.CommandCancel, s.CommandCancelHandler)
	themeCreateCommand := handlers.NewCommand(types.CommandThemeCreate, s.ConversationCreateThemeInit)
	taskCreateCommand := handlers.NewCommand(types.CommandTaskCreateInit, s.ConversationCreateTaskInit)
	themesCommand := handlers.NewCommand(types.CommandThemesGet, s.ConversationThemesInit)
	tasksCommand := handlers.NewCommand(types.CommandTasksGet, s.ConversationTasksInit)
	userEditCommand := handlers.NewCommand(types.CommandUserEdit, s.CommandUserTimezoneHandler)
	allCommands := []handlers.Command{startCommand, cancelCommand, themeCreateCommand, taskCreateCommand, themesCommand, tasksCommand, userEditCommand}

	dispatcher.AddHandler(startCommand)
	dispatcher.AddHandler(cancelCommand)
	//dispatcher.AddHandler(commonCommand)
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal(types.CallbackEmpty), s.CallbackEmpty))

	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{userEditCommand},
		map[string][]ext.Handler{
			types.ConversationUserEditChooseTimezone: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackUserSetTimezone), s.CallbackUserTimezoneChoose),
			},
		}, &handlers.ConversationOpts{
			Exits:        ExitHandlers(userEditCommand, allCommands),
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		},
	))

	//Разговор создания новой темы
	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{themeCreateCommand},
		map[string][]ext.Handler{
			types.ConversationThemeCreateSetName: {handlers.NewMessage(noCommand, s.ConversationCreateThemeSetName)},
		}, &handlers.ConversationOpts{
			Exits:        ExitHandlers(themeCreateCommand, allCommands),
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		}))

	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{tasksCommand},
		map[string][]ext.Handler{
			types.ConversationTaskChoose: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskChoose), s.ConversationTaskChoose),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackChangeTaskPage), s.CallbackTaskChangeTasksPage),
			},
			types.ConversationTaskActionChoose: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskAction), s.ConversationChooseTaskAction),
			},
			//Работа по изменению задачи
			types.ConversationTaskEditSetName: {
				handlers.NewMessage(noCommand, s.ConversationEditTaskSetName),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.CallbackTaskFieldSkip),
			},
			types.ConversationTaskEditSetPriority: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskPrioritySet), s.ConversationEditTaskSetPriority),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.CallbackTaskFieldSkip),
			},
			types.ConversationTaskEditSetDeadline: {
				//Показать клавиатуру выбора срока
				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineShow), s.CallbackDeadlineShowChoose),
				//Выбрать год
				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseYear), s.CallbackTaskDeadlineChooseYear),
				//Выбрать месяц
				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseMonth), s.CallbackTaskDeadlineChooseMonth),
				//Выбрать день
				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseDay), s.CallbackTaskDeadlineChooseDay),
				//Выбрать час
				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseHour), s.CallbackTaskDeadlineChooseHour),
				//Выбрать минуты
				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseMinute), s.CallbackTaskDeadlineChooseMinute),
				//Завершить выбор
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskSetDeadlineDone), s.CallbackTaskDoneDeadline),
				//Пропуск поля
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.CallbackTaskFieldSkip),
			},
			types.ConversationTaskEditSetTheme: {
				//Установка темы для задачи
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskSetTheme), s.ConversationEditTaskSetTheme),
				//Удаление темы для задачи
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskUnsetTheme), s.ConversationEditTaskUnsetTheme),
				//Завершение выбора тем
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskSetThemeDone), s.CallbackTaskDoneTheme),
				//Смена страницы тем
				handlers.NewCallback(callbackquery.Prefix(types.CallbackChangeThemeForTaskPage), s.CallbackTaskChangeThemesPage),
			},

			//Работа по установки статуса задачи
			types.ConversationTaskSetStatus: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskStatusSet), s.ConversationTaskStatusSet),
			},

			//Работа по удалению задачи
			types.ConversationTaskDelete: {
				handlers.NewCallback(callbackquery.Equal(types.CallbackConfirmDelete), s.CallbackTaskDeleteConfirm),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackBackToObject), s.CallbackBackToTask),
			},
		},
		&handlers.ConversationOpts{
			Exits: append(
				ExitHandlers(tasksCommand, allCommands),
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskComplete), s.CallbackTaskDone),
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskStop), s.CallbackTaskCancel),
			),
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		}))

	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{themesCommand},
		map[string][]ext.Handler{
			types.ConversationThemeChoose: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackThemeChoose), s.ConversationThemeChose),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackChangeThemePage), s.CallbackThemeChangeThemesPage),
			},
			types.ConversationThemeActionChoose: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackThemeAction), s.ConversationChoseThemeAction),
			},
			types.ConversationThemeEditSetName: {
				handlers.NewMessage(noCommand, s.ConversationEditThemeSetName),
			},
			types.ConversationThemeDelete: {
				handlers.NewCallback(callbackquery.Equal(types.CallbackConfirmDelete), s.CallbackThemeDeleteConfirm),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackBackToObject), s.CallbackBackToTheme),
			},
		},
		&handlers.ConversationOpts{
			Exits:        ExitHandlers(themesCommand, allCommands),
			AllowReEntry: true,
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		}))

	//Разговор создания новой задачи
	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{taskCreateCommand},
		map[string][]ext.Handler{
			types.ConversationTaskCreateSetName: {
				handlers.NewMessage(noCommand, s.ConversationCreateTaskSetName),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.CallbackTaskFieldSkip),
			},
			types.ConversationTaskCreateSetPriority: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskPrioritySet), s.ConversationCreateTaskSetPriority),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.CallbackTaskFieldSkip),
			},
			types.ConversationTaskCreateSetDeadline: {

				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineShow), s.CallbackDeadlineShowChoose),

				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseYear), s.CallbackTaskDeadlineChooseYear),

				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseMonth), s.CallbackTaskDeadlineChooseMonth),

				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseDay), s.CallbackTaskDeadlineChooseDay),

				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseHour), s.CallbackTaskDeadlineChooseHour),

				handlers.NewCallback(callbackquery.Prefix(types.CallbackDeadlineChooseMinute), s.CallbackTaskDeadlineChooseMinute),

				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskSetDeadlineDone), s.CallbackTaskDoneDeadline),

				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskFieldSkip), s.CallbackTaskFieldSkip),
			},
			types.ConversationTaskCreateSetTheme: {
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskSetTheme), s.ConversationCreateTaskSetTheme),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackTaskUnsetTheme), s.ConversationCreateTaskUnsetTheme),
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskSetThemeDone), s.CallbackTaskDoneTheme),
				handlers.NewCallback(callbackquery.Prefix(types.CallbackChangeThemeForTaskPage), s.CallbackTaskChangeThemesPage),
			},
		},
		&handlers.ConversationOpts{
			Exits: append(
				ExitHandlers(taskCreateCommand, allCommands),
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskComplete), s.CallbackTaskDone),
				handlers.NewCallback(callbackquery.Equal(types.CallbackTaskStop), s.CallbackTaskCancel),
			),
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
		ChatId:   ctx.EffectiveSender.ChatId,
		TimeZone: "Europe/Moscow",
	}
	if err := s.Repository.CreateUser(newUser); err != nil {
		return fmt.Errorf("создание нового пользователя: %w", err)
	}
	if _, err := b.SendMessage(ctx.EffectiveSender.ChatId, fmt.Sprintf("Пользователь создан, по умолчанию используется московское время. Используйте '/%s' для смены", types.CommandUserEdit), nil); err != nil {
		return fmt.Errorf("отправка стартового сообщения: %w", err)
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

func (s *Service) CallbackEmpty(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, nil); err != nil {
		return fmt.Errorf("ответ на выбор пустого обработчика: %w", err)
	}
	return nil
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
	loc, err := time.LoadLocation(user.TimeZone)
	if err != nil {
		return fmt.Errorf("загрузка локации")
	}
	timeNow := time.Now().In(loc)
	task1 := types.TaskModel{
		User:     user,
		Name:     "Первая",
		Deadline: timeNow.Add(time.Hour * 2),
		Status:   types.TaskStatusInWork,
	}
	task2 := types.TaskModel{
		User:     user,
		Name:     "вторая",
		Deadline: timeNow.Add(time.Minute * 1),
		Status:   types.TaskStatusInWork,
	}
	task3 := types.TaskModel{
		User:     user,
		Name:     "третья",
		Deadline: timeNow.Add(time.Hour * 9),
		Status:   types.TaskStatusInWork,
	}
	if _, err := s.Repository.CreateTask(task1); err != nil {
		return fmt.Errorf("создание задачи 1: %w", err)
	}
	if _, err := s.Repository.CreateTask(task2); err != nil {
		return fmt.Errorf("создание задачи 2: %w", err)
	}
	if _, err := s.Repository.CreateTask(task3); err != nil {
		return fmt.Errorf("создание задачи 3: %w", err)
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
	case types.MessageRegisterOperationUserEdit:
		title = "Изменение пользователя прервано"
	case types.MessageRegisterOperationTask:
		title = "Работа с задачей прервана"
	case types.MessageRegisterOperationTheme:
		title = "Работа с темой прервана"
	case types.MessageRegisterOperationTaskStatus:
		title = "Установка статуса задачи прервано"
	case types.MessageRegisterOperationTaskDelete:
		title = "Удаление задачи прервано"
	case types.MessageRegisterOperationThemeDelete:
		title = "Удаление темы прервано"
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

func ExitHandlers(commandException handlers.Command, handlers []handlers.Command) []ext.Handler {
	var exits []ext.Handler
	for _, handler := range handlers {
		if handler.Command != commandException.Command {
			exits = append(exits, handler)
		}
	}
	return exits
}
