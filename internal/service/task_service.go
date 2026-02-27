package service

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"strings"
	"tg-todo/internal/types"
	"tg-todo/internal/utils"
	"time"
)

// ConversationCreateTaskInit - обработчик разговора инициализации создания задачи
func (s *Service) ConversationCreateTaskInit(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := b.SendMessage(ctx.EffectiveSender.ChatId, "Введите имя задачи", nil); err != nil {
		return fmt.Errorf("ошибка отправки сообщения получения имени: %v", err)
	}
	return handlers.NextConversationState(utils.ConversationTaskCreateName)
}

// ConversationCreateTaskSetName - обработчик разговора получения имени задачи
func (s *Service) ConversationCreateTaskSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	user, err := s.Repository.GetUserByTGId(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("ошибка получения пользователя: %v", err)
	}
	newTask := types.TaskModel{
		Name:     ctx.EffectiveMessage.Text,
		User:     user,
		Editable: true,
	}
	if err := s.Repository.CreateTask(newTask); err != nil {
		return fmt.Errorf("ошибка создания задачи: %v", err)
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, "Выберите приоритет задачи", &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.PriorityButtons(),
		},
	}); err != nil {
		return fmt.Errorf("ошибка отправки сообщения получения приоритета: %v", err)
	}
	return handlers.NextConversationState(utils.ConversationTaskCreatePriority)
}

// ConversationCreateTaskSetPriority - обработчик разговора получения приоритета задачи
func (s *Service) ConversationCreateTaskSetPriority(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	priority := types.GetPriorityByCallbackData(callQuery.Data)
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Приоритет выбран"}); err != nil {
		return fmt.Errorf("ошибка ответа на нажатие кнопки: %v", err)
	}
	if _, _, err := callQuery.Message.EditText(b, fmt.Sprintf("Выбран приоритет задачи: %s", priority.Name), nil); err != nil {
		return fmt.Errorf("ошибка редактирования сообщения выбора приоритета: %v", err)
	}
	lastTask, err := s.Repository.GetLastEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("не удалось получить последнюю редактируемую задачу: %v", err)
	}
	lastTask.Priority = priority.Value
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("не удалось обновить приоритет у задачи: %v", err)
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, fmt.Sprintf("Введите сроки задачи в следующем формате %s", utils.TimeLayout), nil); err != nil {
		return fmt.Errorf("ошибка отправки сообщение получения сроков: %v", err)
	}
	return handlers.NextConversationState(utils.ConversationTaskCreateDeadline)
}

// ConversationCreateTaskSetDeadline - обработчик разговора получения сроков задачи
func (s *Service) ConversationCreateTaskSetDeadline(b *gotgbot.Bot, ctx *ext.Context) error {
	deadline, err := time.Parse(utils.TimeLayout, ctx.EffectiveMessage.Text)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf("ошибка обработки времени, введите время в следующем формате %s", utils.TimeLayout), &gotgbot.SendMessageOpts{})
		return handlers.NextConversationState(utils.ConversationTaskCreateDeadline)
	}
	lastTask, err := s.Repository.GetLastEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("не удалось получить последнюю редактируемую задачу: %v", err)
	}
	lastTask.Deadline = deadline
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("не удалось обновить сроки у задачи: %v", err)
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, "Введите темы для задачи через запятую", nil); err != nil {
		return fmt.Errorf("ошибка отправки сообщение получения тем: %v", err)
	}
	return handlers.NextConversationState(utils.ConversationTaskCreateTheme)
}

// ConversationCreateTaskSetTheme -обработчик разговора получения тем задачи
func (s *Service) ConversationCreateTaskSetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	themesNames := strings.Split(ctx.EffectiveMessage.Text, ",")
	for i := range themesNames {
		themesNames[i] = strings.TrimSpace(themesNames[i])
		themesNames[i] = utils.FirstTitleLetter(themesNames[i])
	}
	themes, err := s.Repository.GetThemes(types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, Names: themesNames})
	if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf("ошибка поиска указанных тем, введите их через запятую: %v", err), nil)
		return handlers.NextConversationState(utils.ConversationTaskCreateTheme)
	}
	lastTask, err := s.Repository.GetLastEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("не удалось получить последнюю редактируемую задачу: %v", err)
	}
	if err = s.Repository.UpdateTaskThemes(lastTask, themes); err != nil {
		return fmt.Errorf("ошибка установки тем для задачи: %v", err)
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, "Задача создана", nil); err != nil {
		return fmt.Errorf("ошибка отправки сообщения созданной задачи: %v", err)
	}
	return handlers.EndConversation()
}

// CommandGetTasks - обработчик команды получения задач
func (s *Service) CommandGetTasks(b *gotgbot.Bot, ctx *ext.Context) error {
	tasks, err := s.Repository.GetTasks(types.TaskFilter{UserTGId: ctx.EffectiveSender.User.Id})
	if err != nil {
		return fmt.Errorf("ошибка получения задач: %v", err)
	}
	var taskStroke []string
	for _, task := range tasks {
		var themeStroke []string
		for _, theme := range task.Themes {
			themeStroke = append(themeStroke, fmt.Sprintf("%s", theme.Name))
		}
		priority := types.GetPriorityByValue(task.Priority)
		stroke := fmt.Sprintf("Задача: %s. Приоритет: %s. Срок %s. Темы: %s.", task.Name, priority.Name, task.Deadline.Format(utils.TimeLayout), strings.Join(themeStroke, ", "))
		taskStroke = append(taskStroke, stroke)
	}
	var message string
	if len(taskStroke) > 0 {
		message = strings.Join(taskStroke, "\n")
	} else {
		message = "Нет задач"
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, message, nil); err != nil {
		return fmt.Errorf("ошибка отправки задач: %v", err)
	}
	return nil
}
