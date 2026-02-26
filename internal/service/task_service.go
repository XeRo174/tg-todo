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

//todo выбрать использовать Reply и использовать ответ на сообщение или SendMessage для чистого сообщения

// ConversationCreateTaskInit - обработчик разговора инициализации создания задачи
func (s *Service) ConversationCreateTaskInit(b *gotgbot.Bot, ctx *ext.Context) error {
	if _, err := b.SendMessage(ctx.EffectiveSender.ChatId, "Введите имя задачи", nil); err != nil {
		return fmt.Errorf("ошибка отправки сообщения получения имени: %v", err)
	}
	//if _, err := ctx.EffectiveMessage.Reply(b, "Введите имя задачи", &gotgbot.SendMessageOpts{}); err != nil {
	//	return fmt.Errorf("ошибка отправки сообщения получения имени: %v", err)
	//}
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
	//if _, err = ctx.EffectiveMessage.Reply(b, "Выберите приоритет задачи", &gotgbot.SendMessageOpts{
	//	ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
	//		InlineKeyboard: utils.PriorityButtons(),
	//	},
	//}); err != nil {
	//	return fmt.Errorf("ошибка отправки сообщение получения приоритета: %v", err)
	//}
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
	//if _, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Введите сроки задачи в следующем формате %s", utils.TimeLayout), &gotgbot.SendMessageOpts{}); err != nil {
	//	return fmt.Errorf("ошибка отправки сообщение получения сроков: %v", err)
	//}
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
	//if _, err := ctx.EffectiveMessage.Reply(b,  "Введите темы для задачи через запятую", nil); err != nil {
	//	return fmt.Errorf("ошибка отправки сообщение получения тем: %v", err)
	//}
	return handlers.NextConversationState(utils.ConversationTaskCreateTheme)
}

// ConversationCreateTaskSetTheme -обработчик разговора получения тем задачи
func (s *Service) ConversationCreateTaskSetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	themes, err := s.Repository.GetThemes(ctx.EffectiveSender.User.Id)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf("ошибка поиска указанных тем, введите их через запятую: %v", err), nil)
		return handlers.NextConversationState(utils.ConversationTaskCreateTheme)
	}
	lastTask, err := s.Repository.GetLastEditable(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("не удалось получить последнюю редактируемую задачу: %v", err)
	}
	lastTask.Themes = themes
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("ошибка установки тем для задачи: %v", err)
	}
	b.SendMessage(ctx.EffectiveSender.ChatId, "Задача создана", nil)
	//ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Задача создана"), &gotgbot.SendMessageOpts{})
	return handlers.EndConversation()
}

// CommandGetTasks - обработчик команды получения задач
func (s *Service) CommandGetTasks(b *gotgbot.Bot, ctx *ext.Context) error {
	tasks, err := s.Repository.GetTasks(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("ошибка получения задач: %v", err)
	}
	var taskStroke []string
	for _, task := range tasks {
		var themeStroke []string
		for _, theme := range task.Themes {
			themeStroke = append(themeStroke, fmt.Sprintf("Тема: %s", theme.Name))
		}
		taskStroke = append(taskStroke, fmt.Sprintf("Задача: %s, темы: %s", task.Name, strings.Join(themeStroke, ", ")))
	}
	if _, err = b.SendMessage(ctx.EffectiveSender.ChatId, "Введите темы для задачи через запятую", nil); err != nil {
		return fmt.Errorf("ошибка отправки сообщение получения тем: %v", err)
	}

	if _, err = ctx.EffectiveMessage.Reply(b, strings.Join(taskStroke, "\n"), &gotgbot.SendMessageOpts{}); err != nil {
		return fmt.Errorf("ошибка получения списка задач:  %v", err)
	}
	return nil
}

//// createTask - обработчик инициализации создания задачи
//func (s *Service) createTask(b *gotgbot.Bot, ctx *ext.Context) error {
//	_, err := ctx.EffectiveMessage.Reply(b, "Введите имя задачи", &gotgbot.SendMessageOpts{})
//	if err != nil {
//		return fmt.Errorf("ошибка отправки сообщения получения имени: %v", err)
//	}
//	return handlers.NextConversationState(utils.ConversationTaskCreateName)
//}
//
//// setTaskName - обработчик получения имени задачи и записи в бд
//func (s *Service) setTaskName(b *gotgbot.Bot, ctx *ext.Context) error {
//	user, err := s.GetUserByTGId(ctx.EffectiveSender.User.Id)
//	if err != nil {
//		return err
//	}
//	newTask := types.TaskModel{
//		Name:     ctx.EffectiveMessage.Text,
//		User:     user,
//		Editable: true,
//	}
//	if err := s.Database.Create(&newTask).Error; err != nil {
//		return fmt.Errorf("ошибка создания задачи: %v", err)
//	}
//	var buttons [][]gotgbot.InlineKeyboardButton
//	for _, priority := range types.Priorities {
//		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
//			{Text: priority.Name, CallbackData: fmt.Sprintf("task_priority:%d", priority.Value)},
//		})
//	}
//	if _, err = ctx.EffectiveMessage.Reply(b, "Выберите приоритет задачи", &gotgbot.SendMessageOpts{
//		ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
//			InlineKeyboard: buttons,
//		},
//	}); err != nil {
//		return fmt.Errorf("ошибка отправки сообщение получения приоритета: %v", err)
//	}
//	return handlers.NextConversationState(utils.ConversationTaskCreatePriority)
//}
//// setTaskPriority - обработчик получения приоритета задачи и записи в бд
//func (s *Service) setTaskPriority(b *gotgbot.Bot, ctx *ext.Context) error {
//	callQuery := ctx.Update.CallbackQuery
//	priority := types.GetPriorityByCallbackData(callQuery.Data)
//	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: fmt.Sprintf("Вы выбрали приоритет %s", priority.Name)}); err != nil {
//		return fmt.Errorf("setTaskPriority 1 error: %v", err)
//	}
//	if _, _, err := callQuery.Message.EditText(b, fmt.Sprintf("Выбран приоритет задачи: %s", priority.Name), nil); err != nil {
//		return fmt.Errorf("setTaskPriority 2 error: %v", err)
//	}
//	if err := s.Database.Model(&types.TaskModel{}).
//		Joins("user_models").Where("user_models.tg_id=?", ctx.EffectiveSender.User.Id).Where("task_models.editable=?", true).Update("priority", priority.Value).Error; err != nil {
//		return fmt.Errorf("ошибка установки значения приоритет для задачи: %v", err)
//	}
//	if _, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Введите сроки задачи в следующем формате %v", TimeLayout), &gotgbot.SendMessageOpts{}); err != nil {
//		return fmt.Errorf("ошибка отправки сообщение получения сроков: %v", err)
//	}
//	return handlers.NextConversationState(utils.ConversationTaskCreateDeadline)
//}
//// setTaskDeadline - обработчик получения сроков задачи и записи в бд
//func (s *Service) setTaskDeadline(b *gotgbot.Bot, ctx *ext.Context) error {
//	deadline, err := time.Parse(TimeLayout, ctx.EffectiveMessage.Text)
//	if err != nil {
//		ctx.EffectiveMessage.Reply(b, fmt.Sprintf("ошибка обработки времени, введите время в следующем формате %s", TimeLayout), &gotgbot.SendMessageOpts{})
//		return handlers.NextConversationState(utils.ConversationTaskCreateDeadline)
//	}
//	if err := s.Database.Joins("Join user_models on user_models.id=task_models.user_id").Where("user_models.tg_id=?", ctx.EffectiveSender.User.Id).Where("task_models.editable=?", true).Update("deadline", deadline).Error; err != nil {
//		return fmt.Errorf("ошибка установки значения статуса для задачи: %v", err)
//	}
//	if _, err = ctx.EffectiveMessage.Reply(b, "Введите темы для задачи через запятую", &gotgbot.SendMessageOpts{}); err != nil {
//		return fmt.Errorf("ошибка отправки сообщение получения тем: %v", err)
//	}
//	return handlers.NextConversationState(utils.ConversationTaskCreateTheme)
//}
//// setTaskTheme - обработчик получения темы или тем задачи и записи в бд
//func (s *Service) setTaskTheme(b *gotgbot.Bot, ctx *ext.Context) error {
//	var themes []types.TaskThemeModel
//	if err := s.Database.Joins("join user_models on user_models.id=theme_models.user_id and user_models.tg_id = ?", ctx.EffectiveSender.User.Id).Where("task_theme_models.name in ?", strings.Split(ctx.EffectiveMessage.Text, ",")).Find(&themes).Error; err != nil {
//		ctx.EffectiveMessage.Reply(b, fmt.Sprintf("ошибка поиска указанных тем, введите их через запятую: %v", err), &gotgbot.SendMessageOpts{})
//		return handlers.NextConversationState(utils.ConversationTaskCreateTheme)
//	}
//	if err := s.Database.Joins("join user_models on user_models.id=task_models.user_id").Where("user_models.tg_id=?", ctx.EffectiveSender.User.Id).Where("task_models.editable=?", true).Updates(types.TaskModel{Themes: themes, Editable: false}).Error; err != nil {
//		return fmt.Errorf("ошибка установки тем для задачи: %v", err)
//	}
//	ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Задача создана"), &gotgbot.SendMessageOpts{})
//	return handlers.EndConversation()
//}
//func (s *Service) getTasks(b *gotgbot.Bot, ctx *ext.Context) error {
//	var tasks []types.TaskModel
//	if err := s.Database.Preload("Themes").Joins("join user_models on user_models.id=task_models.user_id and user_models.tg_id = ? ", ctx.EffectiveSender.User.Id).Find(&tasks).Error; err != nil {
//		ctx.EffectiveMessage.Reply(b, fmt.Sprintf("ошибка получения задач: %v", err), &gotgbot.SendMessageOpts{})
//		return handlers.NextConversationState(utils.ConversationTaskCreateTheme)
//	}
//	var taskStroke []string
//	for _, task := range tasks {
//		var themeStroke []string
//		for _, theme := range task.Themes {
//			themeStroke = append(themeStroke, fmt.Sprintf("Тема: %s", theme.Name))
//		}
//		taskStroke = append(taskStroke, fmt.Sprintf("Задача: %s, темы: %s", task.Name, strings.Join(themeStroke, ", ")))
//	}
//	if _, err := ctx.EffectiveMessage.Reply(b, strings.Join(taskStroke, "\n"), &gotgbot.SendMessageOpts{}); err != nil {
//		return fmt.Errorf("getTasks error:  %v", err)
//	}
//	return nil
//}
