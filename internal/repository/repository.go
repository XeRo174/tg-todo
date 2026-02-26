package repository

import (
	"gorm.io/gorm"
	"tg-todo/internal/types"
)

type Repository struct {
	Database *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Database: db,
	}
}

func (r *Repository) GetTasks(userTGId int64) ([]types.TaskModel, error) {
	var tasks []types.TaskModel
	err := r.Database.
		Joins("left join user_models on user_models.id = tasks.user_id").
		Preload("Themes", "deleted_at IS NULL").
		Where("user_models.user_id = ?", userTGId).
		Find(&tasks).Error
	return tasks, err
}

func (r *Repository) GetLastEditable(userTGId int64) (types.TaskModel, error) {
	var task types.TaskModel
	err := r.Database.
		Model(&task).
		Joins("left join users_models on user_models.id = task_models.user_id").
		Where("task_models.editable = true").
		Updates(&task).Error
	return task, err

}

func (r *Repository) CreateTask(task types.TaskModel) error {
	return r.Database.Create(&task).Error
}

func (r *Repository) UpdateTask(task types.TaskModel) error {
	return r.Database.Model(types.TaskModel{}).Where("id = ?", task.ID).Updates(task).Error
}

func (r *Repository) GetThemes(userTGId int64) ([]types.TaskThemeModel, error) {
	var themes []types.TaskThemeModel
	err := r.Database.
		Joins("left join user_models on user_models.id = task_theme_models.user_id").
		Where("user_models.tg_id = ?", userTGId).
		Find(&themes).Error
	return themes, err
}

func getDeletedAtCondition(deletedAt string) (condition string) {
	if deletedAt == "show" {
		condition = ""
	} else if deletedAt == "only" {
		condition = "deleted_at IS NOT NULL"
	} else {
		condition = "deleted_at IS NULL"
	}
	return condition
}
