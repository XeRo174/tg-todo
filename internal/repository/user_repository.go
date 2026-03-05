package repository

import (
	"tg-todo/internal/types"
)

func (r *Repository) CreateUser(user types.UserModel) error {
	return r.Database.Create(&user).Error
}

func (r *Repository) GetUserByTGId(tgId int64) (types.UserModel, error) {
	var user types.UserModel
	if err := r.Database.
		Preload("Messages").
		Preload("Messages.Theme").
		Preload("Messages.Task").
		Preload("Messages.Task.Themes").
		Where("tg_id=?", tgId).First(&user).Error; err != nil {
		return types.UserModel{}, err
	}
	return user, nil
}
