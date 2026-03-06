package types

// Список констант определяющих этапы разговоров
const (
	ConversationThemeCreateInit    = "conversation_theme_create_init"
	ConversationThemeCreateSetName = "conversation_theme_create_set_name"

	ConversationThemeEditInit        = "conversation_theme_edit_init"
	ConversationThemeEditChooseTheme = "conversation_theme_edit_choose_theme"
	ConversationThemeEditSetName     = "conversation_theme_edit_set_name"

	ConversationTaskCreateInit         = "conversation_task_create_init"
	ConversationTaskCreateSetName      = "conversation_task_create_set_name"
	ConversationTaskCreateSetPriority  = "conversation_task_create_set_priority"
	ConversationTaskCreateSetDeadline  = "conversation_task_create_set_deadline"
	ConversationTaskCreateSetTheme     = "conversation_task_create_set_theme"
	ConversationTaskCreateSetThemeDone = "conversation_task_create_set_theme_done"

	ConversationTaskEditInit         = "conversation_task_edit_init"
	ConversationTaskEditChooseTask   = "conversation_task_edit_choose_task"
	ConversationTaskEditSetName      = "conversation_task_edit_set_name"
	ConversationTaskEditSetPriority  = "conversation_task_edit_set_priority"
	ConversationTaskEditSetDeadline  = "conversation_task_edit_set_deadline"
	ConversationTaskEditSetTheme     = "conversation_task_edit_set_theme"
	ConversationTaskEditSetThemeDone = "conversation_task_edit_set_theme_done"
)

// Список констант определяющих ключи callback
const (
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

	// CallbackChangePage - смена страницы
	CallbackChangePage = "change_page:"
)

// Список констант определяющих команды
const (
	CommandThemeCreateInit = "create_theme"
	CommandThemeEditInit   = "edit_theme"
	CommandTaskCreateInit  = "create_task"
	CommandTaskEditInit    = "edit_task"
)

const (
	TimeLayout        = "2006-01-02 15:04:05"
	ThemeKeyboardSize = 3
	UnlimitedSize     = -1

	MessageRegisterOperationTaskCreate = "create_task"
	MessageRegisterOperationTaskEdit   = "edit_task"

	MessageRegisterOperationThemeCreate = "create_theme"
	MessageRegisterOperationThemeEdit   = "edit_theme"
)

const (
	ErrorStrokeFindUserByTG = "поиск пользователя по tg"
)
