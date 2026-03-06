package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"tg-todo/internal/types"
)

func (r *Repository) UpsertMessageRegister(message types.MessageRegisterModel) error {
	return r.Database.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"bot_message_id", "task_id", "theme_id", "operation", "updated_at"}),
	}).Create(&message).Error
}

func (r *Repository) GetMessageRegisterByUserTGId(userTGId int64) (types.MessageRegisterModel, error) {
	var msg types.MessageRegisterModel
	query := r.Database.Model(&types.MessageRegisterModel{})
	query = HandleMessageRegisterPreload(query)
	query = query.Where("user_models.tg_id=?", userTGId)
	err := query.Find(&msg).Error
	return msg, err
}

func (r *Repository) DeleteMessageRegisterByUserId(userId uint) error {
	return r.Database.Unscoped().Model(&types.MessageRegisterModel{}).Where("message_register_models.user_id=?", userId).Delete(&types.MessageRegisterModel{}).Error
}

func HandleMessageRegisterPreload(queryOld *gorm.DB) *gorm.DB {
	queryNew := queryOld
	queryNew = queryNew.Joins("left join user_models on user_models.id = message_register_models.user_id")
	queryNew = queryNew.
		Preload("User").
		Preload("Task").
		Preload("Task.Themes").
		Preload("Theme")
	return queryNew
}
