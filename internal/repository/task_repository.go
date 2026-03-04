package repository

import (
	"fmt"
	"gorm.io/gorm"
	"math"
	"tg-todo/internal/types"
)

// GetTasks - получить задачи с учетом фильтра
func (r *Repository) GetTasks(filter types.TaskFilter) ([]types.TaskModel, error) {
	var tasks []types.TaskModel
	query := r.Database.Model(&types.TaskModel{})
	query = handleTaskPreload(query)
	query = handleTaskFilters(query, filter)
	query = setPagination(query, filter.Size, filter.Page)
	err := query.Find(&tasks).Error
	return tasks, err
}

func (r *Repository) GetTaskPages(filter types.TaskFilter) (float64, error) {
	var count int64
	query := r.Database.Model(&types.TaskModel{})
	query = handleTaskPreload(query)
	query = handleTaskFilters(query, filter)
	err := query.Count(&count).Error
	return math.Ceil(float64(count) / float64(filter.Size)), err
}

// GetLastTaskDraft - получить последнюю созданную задачу черновик
func (r *Repository) GetLastTaskDraft(userTGId int64) (types.TaskModel, error) {
	var task types.TaskModel
	query := r.Database.Model(&types.TaskModel{})
	query = handleTaskPreload(query)
	query = query.
		Where("user_models.tg_id = ?", userTGId).
		Where("task_models.status=?", types.TaskStatusDraft).
		Where("task_message_register_models.operation=?", types.MessageRegisterOperationCreate)
	query = query.Order("task_models.updated_at desc")
	err := query.Last(&task).Error
	return task, err
}

// GetLastTaskEditable - получить последнюю редактируемую задачу
func (r *Repository) GetLastTaskEditable(userTGId int64) (types.TaskModel, error) {
	var task types.TaskModel
	query := r.Database.Model(&types.TaskModel{})
	query = handleTaskPreload(query)
	query = query.
		Where("user_models.tg_id = ?", userTGId).
		Where("task_message_register_models.operation=?", types.MessageRegisterOperationEdit)
	query = query.Order("task_models.updated_at desc")
	err := query.Last(&task).Error
	return task, err
}

func (r *Repository) GetTaskById(userTGId int64, taskId uint) (types.TaskModel, error) {
	var task types.TaskModel
	query := r.Database.Model(&types.TaskModel{})
	query = handleTaskPreload(query)
	query = query.
		Where("task_models.id=?", taskId).
		Where("user_models.tg_id=?", userTGId)
	err := query.First(&task).Error
	return task, err
}

// CreateTask - создать задачу
func (r *Repository) CreateTask(task types.TaskModel) (types.TaskModel, error) {
	return task, r.Database.Create(&task).Error
}

// UpdateTask - обновить данные задачи
func (r *Repository) UpdateTask(task types.TaskModel) error {
	return r.Database.Model(types.TaskModel{}).Where("id = ?", task.ID).Updates(task).Error
}

// UpdateTaskThemes - обновить темы задачи
func (r *Repository) UpdateTaskThemes(task types.TaskModel, themes []types.ThemeModel) error {
	return r.Database.Model(&task).Association("Themes").Replace(themes)
}

// DeleteTask - удалить задачу
func (r *Repository) DeleteTask(id uint) error {
	return r.Database.Delete(&types.TaskModel{}, id).Error
}

func (r *Repository) WriteTaskMessage(taskMessage types.MessageRegisterModel) error {
	return r.Database.Create(&taskMessage).Error
}

// todo условия обновления? message id полностью уникальное? можно ли использовать только его для поиска
func (r *Repository) UpdateTaskMessageByMessageIdAndTaskId(taskMessage types.MessageRegisterModel) error {
	return r.Database.Model(types.MessageRegisterModel{}).
		Joins("left join task_models on task_models.tg_id = task_message_register_models.task_id").
		Where("id=?", taskMessage.ID).
		Where("task_message_register_models.bot_message_id=?", taskMessage.BotMessageId).
		Updates(taskMessage).Error
}

// handleTaskPreload - сформировать запрос подключения внешних таблиц к таблице задач и предварительно загрузить данные
func handleTaskPreload(query *gorm.DB) *gorm.DB {
	query = query.
		Joins("left join user_models on user_models.id = task_models.user_id").
		Joins("left join task_themes on task_themes.theme_model_id = task_models.id").
		Joins("left join theme_models on theme_models.id = task_themes.theme_model_id").
		Joins("left join task_message_register_models on task_message_register_models.task_id=task_models.id")

	query = query.
		Preload("Themes").
		Preload("Messages")

	return query
}

// handleTaskFilters - сформировать запрос проверки условий таблицы задач
func handleTaskFilters(queryOld *gorm.DB, filter types.TaskFilter) *gorm.DB {
	queryNew := queryOld
	if filter.Name != "" {
		condition, value := wrapCondition(filter.LikeSearch, filter.Name)
		queryNew = queryNew.Where(fmt.Sprintf("task_models.name"+condition), value)
	}
	if len(filter.Names) > 0 {
		queryNew = queryNew.Where(fmt.Sprintf("task_models.name IN (?) "), filter.Names)
	}
	if filter.Status != 0 {
		queryNew = queryNew.Where(fmt.Sprintf("task_models.status = ?"), filter.Status)
	}
	if filter.Priority != 0 {
		queryNew = queryNew.Where(fmt.Sprintf("task_models.priority = ?"), filter.Priority)
	}
	if !filter.Deadline.IsZero() {
		queryNew = queryNew.Where("task_models.deadline = ?", filter.Deadline)
	}
	if filter.UserId > 0 {
		queryNew = queryNew.Where("task_models.user_id = ?", filter.UserId)
	}
	if filter.UserTGId > 0 {
		queryNew = queryNew.Where("user_models.tg_id = ?", filter.UserTGId)
	}
	if filter.ThemeName != "" {
		condition, value := wrapCondition(filter.LikeSearch, filter.ThemeName)
		queryNew = queryNew.Where(fmt.Sprintf("theme_models.name"+condition), value)
	}
	if len(filter.ThemeNames) > 0 {
		queryNew = queryNew.Where("theme_models.name IN (?)", filter.ThemeNames)
	}
	return queryNew
}
