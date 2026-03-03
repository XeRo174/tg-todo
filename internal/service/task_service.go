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

func TaskMessageFill(title, ending string, task types.TaskModel, themes []types.ThemeModel) string {
	return fmt.Sprintf("%s\nНазвание: %s\nПриоритет: %s\nСроки: %s\nТемы: %s\n\n%s", title, task.Name, task.Priority.String(), task.Deadline.Format(types.TimeLayout), utils.ThemeStroke(themes), ending)
}

// ConversationCreateTaskInit - обработчик разговора инициализации создания задачи
func (s *Service) ConversationCreateTaskInit(b *gotgbot.Bot, ctx *ext.Context) error {
	tasks, err := s.Repository.GetTasks(types.TaskFilter{UserTGId: ctx.EffectiveSender.User.Id})
	if err != nil {
		return fmt.Errorf("получение задач: %w", err)
	}
	user, err := s.Repository.GetUserByTGId(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск пользователя по tg: %w", err)
	}
	newTask := types.TaskModel{
		Name:   fmt.Sprintf("Задача №%d", len(tasks)+1),
		User:   user,
		Status: types.TaskStatusDraft,
	}
	createdTask, err := s.Repository.CreateTask(newTask)
	if err != nil {
		return fmt.Errorf("создание задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", "Введите имя задачи", createdTask, nil)
	taskMessage, err := b.SendMessage(ctx.EffectiveSender.ChatId, message, &gotgbot.SendMessageOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TaskButtons(),
		},
	})
	if err != nil {
		return fmt.Errorf("отправка сообщения задачи: %w", err)
	}
	if err = s.Repository.WriteTaskMessage(types.TaskMessageRegister{TaskId: createdTask.ID, BotMessageId: taskMessage.MessageId}); err != nil {
		return fmt.Errorf("запись сообщения задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationNewTaskName)
}

// ConversationCreateTaskSetName - обработчик разговора получения имени задачи
func (s *Service) ConversationCreateTaskSetName(b *gotgbot.Bot, ctx *ext.Context) error {
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	messageId, ok := utils.TaskMessageExist(lastTask)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	lastTask.Name = ctx.EffectiveMessage.Text
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", "Выберите приоритет", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: append(utils.PriorityButtons(), utils.TaskButtons()...),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения имени задачи")
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
	messageId, ok := utils.TaskMessageExist(lastTask)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	lastTask.Priority = priority
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление приоритета задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Введите срок в формате %s", types.TimeLayout), lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageId,
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
	messageId, ok := utils.TaskMessageExist(lastTask)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	lastTask.Deadline = deadline
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление сроков задачи: %w", err)
	}
	themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: 2, Page: 1}}
	themes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем для клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.CreateThemeInlineKeyboard(lastTask, themes, 1, themesPagesCount),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, сроки задачи: %w", err)
	}
	if _, err = b.DeleteMessage(ctx.EffectiveSender.ChatId, ctx.EffectiveMessage.MessageId, nil); err != nil {
		return fmt.Errorf("удаление сообщения сроков задачи")
	}
	return handlers.NextConversationState(types.ConversationNewTaskThemeChoose)
}

// ConversationCreateTaskSetTheme - обработчик разговора получения тем задачи
func (s *Service) ConversationCreateTaskSetTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Тема выбрана"}); err != nil {
		return fmt.Errorf("ответ на выбор темы: %w", err)
	}
	callQueryValues := strings.Split(callQuery.Data, ";")
	var themeId, currentPage uint
	for _, callQueryData := range callQueryValues {
		if strings.HasPrefix(callQueryData, types.CallbackTaskThemeChoose) {
			themeStr := strings.Replace(callQueryData, types.CallbackTaskThemeChoose, "", 1)
			theme, err := strconv.ParseUint(themeStr, 10, 64)
			if err != nil {
				return fmt.Errorf("получение идентификатора темы из клавиатуры: %w", err)
			}
			themeId = uint(theme)
		}
		if strings.HasPrefix(callQueryData, types.CallbackCurrentPage) {
			pageStr := strings.Replace(callQueryData, types.CallbackCurrentPage, "", 1)
			page, err := strconv.ParseUint(pageStr, 10, 64)
			if err != nil {
				return fmt.Errorf("получение номера страницы клавиатуры тем: %w", err)
			}
			currentPage = uint(page)
		}
	}

	themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: 2, Page: currentPage}}
	allThemes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получение тем для клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	var chosenTheme types.ThemeModel
	found := false
	for _, theme := range allThemes {
		if theme.ID == themeId {
			chosenTheme = theme
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("выбранная тема не найдена")
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	messageId, ok := utils.TaskMessageExist(lastTask)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	lastTask.Themes = append(lastTask.Themes, chosenTheme)
	if err = s.Repository.UpdateTaskThemes(lastTask, lastTask.Themes); err != nil {
		return fmt.Errorf("обновление тем задачи: %w", err)
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.CreateThemeInlineKeyboard(lastTask, allThemes, currentPage, themesPagesCount),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, темы задачи: %w", err)
	}
	return handlers.NextConversationState(types.ConversationNewTaskThemeChoose)
}

func (s *Service) ConversationCreateTaskDoneTheme(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Выбор завершен"}); err != nil {
		return fmt.Errorf("ответ на завершение выбора: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	messageId, ok := utils.TaskMessageExist(lastTask)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	message := TaskMessageFill("Создание задачи", "Нажмите Создать", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: utils.TaskButtons(),
		},
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, завершение создания: %w", err)
	}
	return nil
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
	messageId, ok := utils.TaskMessageExist(lastTask)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	message := TaskMessageFill("Задача создана", "", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, завершение создания: %w", err)
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
	messageId, ok := utils.TaskMessageExist(lastTask)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	lastTask.Status = types.TaskStatusDropped
	if err = s.Repository.UpdateTask(lastTask); err != nil {
		return fmt.Errorf("обновление статуса задачи: %w", err)
	}
	message := TaskMessageFill("Задача отменена", "", lastTask, lastTask.Themes)
	if _, _, err = b.EditMessageText(message, &gotgbot.EditMessageTextOpts{
		MessageId: messageId,
		ChatId:    ctx.EffectiveSender.ChatId,
	}); err != nil {
		return fmt.Errorf("изменение сообщения задачи, отмена создания: %w", err)
	}
	return handlers.EndConversation()
}

func (s *Service) ConversationCreateTaskChangePage(b *gotgbot.Bot, ctx *ext.Context) error {
	callQuery := ctx.Update.CallbackQuery
	if _, err := callQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Страница тем изменена"}); err != nil {
		return fmt.Errorf("ответ на переключение страницы: %w", err)
	}
	pageStr := strings.Replace(callQuery.Data, types.CallbackChangeTaskThemesPage, "", 1)
	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		return fmt.Errorf("получение номера страницы клавиатуры тем: %w", err)
	}
	themeFilter := types.ThemeFilter{UserTGId: ctx.EffectiveSender.User.Id, SortQuery: types.SortQuery{Size: 2, Page: uint(page)}}
	allThemes, err := s.Repository.GetThemes(themeFilter)
	if err != nil {
		return fmt.Errorf("получения тем для клавиатуры: %w", err)
	}
	themesPagesCount, err := s.Repository.GetThemePages(themeFilter)
	if err != nil {
		return fmt.Errorf("получение количества страниц тем: %w", err)
	}
	lastTask, err := s.Repository.GetLastTaskDraft(ctx.EffectiveSender.User.Id)
	if err != nil {
		return fmt.Errorf("поиск последней задачи: %w", err)
	}
	messageId, ok := utils.TaskMessageExist(lastTask)
	if !ok {
		return fmt.Errorf("сообщение задачи не найдено")
	}
	message := TaskMessageFill("Создание задачи", fmt.Sprintf("Выберите темы.\nСтраница: (%d/%v)", themeFilter.Page, themesPagesCount), lastTask, lastTask.Themes)
	if _, _, err = callQuery.Message.EditText(b, message, &gotgbot.EditMessageTextOpts{
		MessageId: messageId,
		ChatId:    ctx.EffectiveSender.ChatId,
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: s.CreateThemeInlineKeyboard(lastTask, allThemes, uint(page), themesPagesCount),
		},
	}); err != nil {
		return fmt.Errorf("изменение страницы клавиатуры тем: %w", err)
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

func (s *Service) CreateThemeInlineKeyboard(lastTask types.TaskModel, themesByPage []types.ThemeModel, page uint, pagesCount float64) [][]gotgbot.InlineKeyboardButton {
	var buttons [][]gotgbot.InlineKeyboardButton
	var chosenThemes []types.ThemeModel
	for _, theme := range themesByPage {
		var buttonText string
		if utils.Contains(lastTask.Themes, theme.Name) {
			buttonText = fmt.Sprintf("[x] %s", theme.Name)
			chosenThemes = append(chosenThemes, theme)
		} else {
			buttonText = fmt.Sprintf("[  ] %s", theme.Name)
		}
		buttons = append(buttons, []gotgbot.InlineKeyboardButton{
			{Text: buttonText, CallbackData: fmt.Sprintf("%s%d;%s%d", types.CallbackTaskThemeChoose, theme.ID, types.CallbackCurrentPage, page)},
		})
	}
	totalPages := int(pagesCount)
	currentPage := int(page)

	if totalPages > 1 {
		if currentPage == 1 {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", totalPages, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, totalPages)},
				{Text: fmt.Sprintf("(%d/%d) >", currentPage+1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage+1)},
			})
		} else if currentPage == totalPages {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", 1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, 1)},
			})
		} else if currentPage == totalPages {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", 1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, 1)},
			})
		} else {
			buttons = append(buttons, []gotgbot.InlineKeyboardButton{
				{Text: fmt.Sprintf("< (%d/%d)", currentPage-1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage-1)},
				{Text: fmt.Sprintf("(%d/%d) >", currentPage+1, totalPages), CallbackData: fmt.Sprintf("%s%d", types.CallbackChangeTaskThemesPage, currentPage+1)},
			})
		}
	}
	buttons = append(buttons, []gotgbot.InlineKeyboardButton{
		{Text: "Завершить", CallbackData: types.CallbackThemeChoseDone},
	})
	return append(buttons, utils.TaskButtons()...)
}
