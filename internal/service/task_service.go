package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"strconv"
	"strings"
	"tg-todo/internal/types"
	"tg-todo/internal/utils"
	"time"
)

var taskMessageId int64

func TaskMessageFill(title, ending string, task types.TaskModel, themes []types.ThemeModel) string {
	return fmt.Sprintf("%s\nНазвание: %s\nПриоритет: %s\nСроки: %s\nТемы: %s\n\n%s", title, task.Name, task.Priority.String(), task.Deadline.Format(types.TimeLayout), utils.ThemeStroke(themes), ending)
}

// ConversationCreateTaskInit - обработчик разговора инициализации создания задачи
func (s *Service) ConversationCreateTaskInit(b *gotgbot.Bot, ctx *ext.Context) error {
	message := TaskMessageFill("Создание задачи", "Введите имя задачи", types.TaskModel{}, nil)
	taskMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TaskButtons(),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения задачи: %w", err)
	}
	//todo task message_id
	taskMessageId = taskMessage.MessageId
	return handlers.NextConversationState(types.ConversationNewTaskName)
}

// ConversationCreateTaskSetName - обработчик разговора получения имени задачи
func (s *Service) ConversationCreateTaskSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := s.Repository.GetUserByTGId(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск пользователя по tg: %w", err)
	}
	newTask := types.TaskModel{
		Name:   ctx.EffectiveMessage.Text,
		User:   user,
		Status: types.TaskStatusDraft,
	}
	if err = s.Repository.CreateTask(newTask); err != nil {
		return fmt.Errorf("создание задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", "Выберите приоритет", newTask, newTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: taskMessageId, //todo task message_id
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.PriorityButtons(), utils.TaskButtons()...),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationNewTaskPriority)
}

// ConversationCreateTaskSetPriority - обработчик разговора получения приоритета задачи
func (s *Service) ConversationCreateTaskSetPriority(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Приоритет выбран"}); err != nil {
		return fmt.Errorf("ответ на выбор приоритета: %w", err)
	}
	priority, ok := types.ParsePriority(callQuery.Data)
	if !ok {
		priority = types.TaskPriorityNone
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	lastTask.Priority = priority
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление приоритета задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Введите срок в формате %s", types.TimeLayout), lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: taskMessageId, //todo task message_id
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TaskButtons(),
		},
		//todo inline keyboard календарь + время
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, приоритет задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationNewTaskDeadline)
}

// ConversationCreateTaskSetDeadline - обработчик разговора получения сроков задачи
func (s *Service) ConversationCreateTaskSetDeadline(b *gotgbot.Bot, ctx *ext.Context) error {
	deadline, err := time.Parse(types.TimeLayout, ctx.EffectiveMessage.Text)
	if err != nil {
		if _, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf("обработка сроков задачи, используйте формат %s", types.TimeLayout), nil); err != nil {
			return fmt.Errorf("ответ про правильный формат сроков: %w", err)
		}
		return handlers.NextConversationState(types.ConversationNewTaskDeadline)
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	lastTask.Deadline = deadline
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	themes, err := s.Repository.GetThemes(types.ThemeFilter{
		UserTGId: ctx.EffectiveSender.User.Id,
		SortQuery: types.SortQuery{
			Size: 2,
			Page: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("получени тем для клавиатуры: %w", err)
	}
	message := TaskMessageFill("Создание задачи", "Выберите темы", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: taskMessageId, //todo task message_id
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.ThemesButton(themes), utils.TaskButtons()...),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, сроки задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationNewTaskThemeChoose)
}

// ConversationCreateTaskSetTheme -обработчик разговора получения тем задачи
func (s *Service) ConversationCreateTaskSetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	themeName := strings.Replace(callQuery.Data, "set_task_theme:", "", 1)
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор темы: %w", err)
	}
	allThemes, err := s.Repository.GetThemes(types.ThemeFilter{
		UserTGId: ctx.EffectiveSender.User.Id,
		SortQuery: types.SortQuery{
			Size: 2,
			Page: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("получени тем для клавиатуры: %w", err)
	}
	chosenThemes, err := s.Repository.GetThemes(types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, Name: themeName})
	if err != nil {
		return fmt.Errorf("получение выбранной темы: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	if err = s.Repository.UpdateTaskThemes(lastTask, append(lastTask.Themes, chosenThemes...)); err != nil {
		return fmt.Errorf("обновление тем задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", "Выберите темы", lastTask, append(lastTask.Themes, chosenThemes...))
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: taskMessageId, //todo task message_id
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.ThemesButton(allThemes), utils.TaskButtons()...), //все темы надо выводить
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, темы задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationNewTaskThemeChoose)
}

// ConversationCreateTaskDone - завершение создания
func (s *Service) ConversationCreateTaskDone(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Завершение создания"}); err != nil {
		return fmt.Errorf("ответ на завершение создания: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	lastTask.Status = types.TaskStatusCreated
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление статуса задачи: %w", err)
	}
	message := TaskMessageFill("Задача создана", "", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: taskMessageId, //todo task message_id
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, завершение создания: %w", err)
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, "Задача создана", nil); err != nil {
		return fmt.Errorf("отправка сообщения завершения создания: %w", err)
	}
	return handlers.EndConversation()
}

// ConversationCreateTaskCancel - отмена создания
func (s *Service) ConversationCreateTaskCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Отмена создания"}); err != nil {
		return fmt.Errorf("ответ на отмену создания: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	lastTask.Status = types.TaskStatusDropped
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление статуса задачи: %w", err)
	}
	message := TaskMessageFill("Задача отменена", "", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: taskMessageId, //todo task message_id
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, отмена создания: %w", err)
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, "Задача отменена", nil); err != nil {
		return fmt.Errorf("отправка сообщения отмены создания: %w", err)
	}
	return handlers.EndConversation()
}

func (s *Service) ConversationCreateTaskChangePage(b *gotgbot.Bot, ctx *ext.Context) error {
	s.App.Logger.Info("start")
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема выбрана"}); err != nil {
		return fmt.Errorf("ошибка ответа на нажатие кнопки: %w", err)
	}
	pageStr := strings.Replace(callQuery.Data, "change_theme_page:", "", 1)
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return fmt.Errorf("не удалось обработать новую страницу темы: %v", err)
	}
	filter := types.ThemeFilter{
		UserTGId: ctx.EffectiveSender.User.Id,
	}
	if page <= 0 {
		filter.Page = 0
	}
	themes, err := s.Repository.GetThemes(filter)
	if err != nil {
		return fmt.Errorf("ошибка получения тем клавиатуры: %v", err)
	}
	if _, _, err = callQuery.Message.EditText(b, fmt.Sprintf("Выберите тему. \nСтраница: %d", filter.Page), nil); err != nil {
		return fmt.Errorf("ошибка редактирования сообщения смены страницы: %v", err)
	}
	if _, _, err = callQuery.Message.EditReplyMarkup(b, &gotgbot.EditMessageReplyMarkupOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.ThemesButton(themes),
		},
	}); err != nil {
		return fmt.Errorf("ошибка редактирования клавиш тем: %v", err)
	}
	return nil

}

// CommandGetTasks - обработчик команды получения задач
func (s *Service) CommandGetTasks(b *gotgbot.Bot, ctx *ext.Context) error {
	tasks, err := s.Repository.GetTasks(types.TaskFilter{UserTGId: ctx.EffectiveSender.User.Id})
	if err != nil {
		return fmt.Errorf("получение задач: %w", err)
	}
	var taskStroke []string
	for _, task := range tasks {
		var themeStroke []string
		for _, theme := range task.Themes {
			themeStroke = append(themeStroke, fmt.Sprintf("%s", theme.Name))
		}
		priority, ok := types.ParsePriority(fmt.Sprintf("%d", task.Priority))
		if !ok {
			priority = types.TaskPriorityNone
		}
		stroke := fmt.Sprintf("Задача: %s. Приоритет: %s. Статус: %s. Срок %s. Темы: %s.", task.Name, priority.String(), task.Status.String(), task.Deadline.Format(types.TimeLayout), strings.Join(themeStroke, ", "))
		taskStroke = append(taskStroke, stroke)
	}
	var message string
	if len(taskStroke) > 0 {
		message = strings.Join(taskStroke, "\n\n")
	} else {
		message = "Нет задач"
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, message, nil); err != nil {
		return fmt.Errorf("отправка задач: %w", err)
	}
	return nil
}
