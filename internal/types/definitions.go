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

	//ConversationNewTaskName        = "conversation_new_task_name"
	//ConversationNewTaskPriority    = "conversation_new_task_priority"
	//ConversationNewTaskDeadline    = "conversation_new_task_deadline"
	//ConversationNewTaskThemeChoose = "conversation_new_task_theme_choose"
)

// Список констант определяющих ключи callback
const (
	CallbackThemeEditChooseTheme      = "theme_edit_choose_theme:"
	CallbackThemeEditChangeThemesPage = "theme_edit_change_themes_page:"

	CallbackThemeChoose = "theme_choose:"

	CallbackTaskChoose    = "task_choose:"
	CallbackTaskFieldSkip = "task_field_skip:"
	CallbackTaskComplete  = "task_complete:"
	CallbackTaskDone      = "task_done:"
	CallbackTaskCancel    = "task_cancel:"

	CallbackTaskPrioritySet      = "set_task_priority:"
	CallbackTaskStatusSet        = "set_task_status:"
	CallbackTaskThemeChoose      = "set_task_theme:"
	CallbackChangeTaskThemesPage = "change_page_task_themes:"
	CallbackCurrentPage          = "current_page:"
	CallbackThemeChoseDone       = "chose_theme_done"
	CallbackTaskCreateDone       = "create_task_done"
	CallbackTaskCreateCancel     = "create_task_cancel:"

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

	MessageRegisterOperationCreate = "create"
	MessageRegisterOperationEdit   = "edit"

	MessageRegisterOperationTaskCreate = "create_task"
	MessageRegisterOperationTaskEdit   = "edit_task"

	MessageRegisterOperationThemeCreate = "create_theme"
	MessageRegisterOperationThemeEdit   = "edit_theme"
)

const (
	ErrorStrokeFindUserByTG = "поиск пользователя по tg"
)
