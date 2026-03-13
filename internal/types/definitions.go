package types

// Список констант определяющих этапы разговоров
const (
	ConversationUserEditInit           = "conversation_user_edit_init"
	ConversationUserEditChooseTimezone = "conversation_user_edit_choose_timezone"

	ConversationThemeCreateInit    = "conversation_theme_create_init"
	ConversationThemeCreateSetName = "conversation_theme_create_set_name"

	ConversationThemeEditInit        = "conversation_theme_edit_init"
	ConversationThemeEditChooseTheme = "conversation_theme_edit_choose_theme"
	ConversationThemeEditSetName     = "conversation_theme_edit_set_name"

	ConversationThemeInit         = "conversation_theme_init"
	ConversationThemeChoose       = "conversation_theme_choose"
	ConversationThemeActionChoose = "conversation_theme_action_choose"

	ConversationThemeDelete = "conversation_theme_delete"

	ConversationTaskInit         = "conversation_task_init"
	ConversationTaskChoose       = "conversation_task_choose"
	ConversationTaskActionChoose = "conversation_task_action_choose"

	ConversationTaskEditSetName     = "conversation_task_edit_set_name"
	ConversationTaskEditSetPriority = "conversation_task_edit_set_priority"
	ConversationTaskEditSetDeadline = "conversation_task_edit_set_deadline"
	ConversationTaskEditSetTheme    = "conversation_task_edit_set_theme"

	ConversationTaskSetStatus = "conversation_task_set_status"

	ConversationTaskDelete = "conversation_task_delete"

	ConversationTaskCreateInit         = "conversation_task_create_init"
	ConversationTaskCreateSetName      = "conversation_task_create_set_name"
	ConversationTaskCreateSetPriority  = "conversation_task_create_set_priority"
	ConversationTaskCreateSetDeadline  = "conversation_task_create_set_deadline"
	ConversationTaskCreateSetTheme     = "conversation_task_create_set_theme"
	ConversationTaskCreateSetThemeDone = "conversation_task_create_set_theme_done"
)

// Список констант определяющих ключи callback
const (
	// CallbackUserSetTimezone - установка временной зоны пользователя
	CallbackUserSetTimezone = "set_user_timezone:"

	// CallbackThemeChoose - выбор темы для редактирования (новый)
	CallbackThemeChoose = "theme_choose:"

	// CallbackTaskChoose - выбор задачи для редактирования
	CallbackTaskChoose = "task_choose:"
	// CallbackTaskFieldSkip - пропуск поля задачи
	CallbackTaskFieldSkip = "task_field_skip:"
	// CallbackTaskComplete - завершение работы с задачей
	CallbackTaskComplete = "task_complete"
	// CallbackTaskStop - прекращение работы с задачей
	CallbackTaskStop = "task_stop"

	// CallbackTaskAction - действие которое надо произвести с задачей
	CallbackTaskAction = "task_action:"

	// CallbackThemeAction - действие которое надо произвести с темой
	CallbackThemeAction = "theme_action:"

	// CallbackTaskPrioritySet - установка приоритета задачи
	CallbackTaskPrioritySet = "set_task_priority:"
	// CallbackTaskStatusSet - установка статуса задачи
	CallbackTaskStatusSet = "set_task_status:"
	// CallbackTaskSetTheme - установка темы задачи
	CallbackTaskSetTheme = "task_set_theme:"
	// CallbackTaskUnsetTheme - удаления темы задачи
	CallbackTaskUnsetTheme = "task_unset_theme:"
	// CallbackTaskSetThemeDone - завершение выбора тем задачи
	CallbackTaskSetThemeDone = "task_set_theme_done"

	// CallbackCurrentPage - текущая страница
	CallbackCurrentPage = "current_page:"

	// CallbackChangeThemePage - смена страницы выбора тем
	CallbackChangeThemePage = "change_theme_page:"
	// CallbackChangeTaskPage - смена страницы выбора задач
	CallbackChangeTaskPage = "change_task_page:"
	// CallbackChangeThemeForTaskPage - смена страницы выбора тем для задачи
	CallbackChangeThemeForTaskPage = "change_theme_for_task_page:"

	// CallbackDeadlineShow - выбор параметров сроков для отображения
	CallbackDeadlineShow = "deadline_show:"

	// CallbackDeadlineChooseYear - установка года сроков
	CallbackDeadlineChooseYear = "deadline_choose_year:"
	// CallbackDeadlineChooseMonth - установка месяца сроков
	CallbackDeadlineChooseMonth = "deadline_choose_month:"
	// CallbackDeadlineChooseDay - установка дня сроков
	CallbackDeadlineChooseDay = "deadline_choose_day:"
	// CallbackDeadlineChooseHour - установка часа сроков
	CallbackDeadlineChooseHour = "deadline_choose_hour:"
	// CallbackDeadlineChooseMinute - установка минут сроков
	CallbackDeadlineChooseMinute = "deadline_choose_minute:"
	// CallbackTaskSetDeadlineDone - завершение установки сроков
	CallbackTaskSetDeadlineDone = "task_set_deadline_done:"

	CallbackConfirmDelete = "delete_confirm"
	CallbackBackToObject  = "back_to_object"

	CallbackEmpty = "callback_empty_skip"
)

// Список констант определяющих команды
const (
	CommandStart          = "start"
	CommandCancel         = "cancel"
	CommandThemeCreate    = "create_theme"
	CommandTaskCreateInit = "create_task"
	CommandThemesGet      = "themes"
	CommandTasksGet       = "tasks"
	CommandUserEdit       = "edit_user"
)

const (
	TimeLayout        = "2006-01-02 15:04"
	ThemeKeyboardSize = 3
	UnlimitedSize     = -1

	MessageRegisterOperationTask  = "task"
	MessageRegisterOperationTheme = "theme"

	MessageRegisterOperationTaskCreate = "create_task"
	MessageRegisterOperationTaskEdit   = "edit_task"
	MessageRegisterOperationTaskStatus = "status_task"
	MessageRegisterOperationTaskDelete = "delete_task"

	MessageRegisterOperationThemeCreate = "create_theme"
	MessageRegisterOperationThemeEdit   = "edit_theme"
	MessageRegisterOperationThemeDelete = "delete_theme"

	MessageRegisterOperationUserEdit = "edit_user"

	ActionEdit   = "action_edit"
	ActionStatus = "action_status"
	ActionDelete = "action_delete"
	ActionBack   = "action_back"

	DeadlineShowYears   = "years"
	DeadlineShowMonths  = "months"
	DeadlineShowDays    = "days"
	DeadlineShowHours   = "hours"
	DeadlineShowMinutes = "minutes"
)

const (
	ErrorStrokeFindUserByTG = "поиск пользователя по tg"
)
