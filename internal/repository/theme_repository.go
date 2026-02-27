package repository

import (
	"fmt"
	"gorm.io/gorm"
	"tg-todo/internal/types"
)

// GetThemes - получить темы с учетом фильтра
func (r *Repository) GetThemes(filter types.ThemeFilter) ([]types.ThemeModel, error) {
	var themes []types.ThemeModel
	query := r.Database.Model(&types.ThemeModel{})
	query = handleThemePreload(query)
	query = handleThemeFilters(query, filter)
	err := query.Find(&themes).Error
	return themes, err
}

// CreateTheme - создать тему
func (r *Repository) CreateTheme(theme types.ThemeModel) error {
	return r.Database.Create(&theme).Error
}

// UpdateTheme - обновить тему
func (r *Repository) UpdateTheme(theme types.ThemeModel) error {
	return r.Database.Model(types.ThemeModel{}).Where("id = ?", theme.ID).Updates(theme).Error
}

// handleTaskPreload - сформировать запрос подключения внешних таблиц к таблице тем и предварительно загрузить данные
func handleThemePreload(query *gorm.DB) *gorm.DB {
	query = query.
		Joins("left join user_models on user_models.id = theme_models.user_id")
	query = query.Preload("Tasks")
	return query
}

// handleTaskFilters - сформировать запрос проверки условий таблицы тем
func handleThemeFilters(queryOld *gorm.DB, filter types.ThemeFilter) *gorm.DB {
	queryNew := queryOld
	if filter.Name != "" {
		condition, value := wrapCondition(filter.LikeSearch, filter.Name)
		queryNew = queryNew.Where(fmt.Sprintf("theme_models.name"+condition), value)
	}
	if len(filter.Names) > 0 {
		queryNew = queryNew.Where("theme_models.name IN (?)", filter.Names)
	}
	if filter.UserId > 0 {
		queryNew = queryNew.Where("theme_models.user_id = ?", filter.UserId)
	}
	if filter.UserTGId > 0 {
		queryNew = queryNew.Where("user_models.tg_id = ?", filter.UserTGId)
	}
	return queryNew
}
