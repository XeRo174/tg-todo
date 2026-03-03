package types

// Список констант определяющих этапы разговоров
const (
	ConversationNewThemeName = "conversation_new_theme_name"

	ConversationNewTaskName        = "conversation_new_task_name"
	ConversationNewTaskPriority    = "conversation_new_task_priority"
	ConversationNewTaskDeadline    = "conversation_new_task_deadline"
	ConversationNewTaskThemeChoose = "conversation_new_task_theme_choose"
)

// Список констант определяющих ключи callback
const (
	CallbackTaskPrioritySet      = "set_task_priority:"
	CallbackTaskStatusSet        = "set_task_status:"
	CallbackTaskThemeChoose      = "set_task_theme:"
	CallbackChangeTaskThemesPage = "change_page_task_themes:"
	CallbackCurrentPage          = "current_page:"
	CallbackThemeChoseDone       = "chose_theme_done"
	CallbackTaskCreateDone       = "create_task_done"
	CallbackTaskCreateCancel     = "create_task_cancel"
)

const (
	TimeLayout = "2006-01-02 15:04:05"
)
