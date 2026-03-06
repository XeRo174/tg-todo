package repository

import (
	"fmt"
	"gorm.io/gorm"
	"math"
	"tg-todo/internal/types"
)

// GetThemes - получить темы с учетом фильтра
func (r *Repository) GetThemes(filter types.ThemeFilter) ([]types.ThemeModel, error) {
	var themes []types.ThemeModel
	query := r.Database.Model(&types.ThemeModel{})
	query = handleThemePreload(query)
	query = handleThemeFilters(query, filter)
	query = setPagination(query, filter.Size, filter.Page)
	err := query.Find(&themes).Error
	return themes, err
}

func (r *Repository) GetThemePages(filter types.ThemeFilter) (float64, error) {
	var count int64
	query := r.Database.Model(&types.ThemeModel{})
	query = handleThemePreload(query)
	query = handleThemeFilters(query, filter)
	err := query.Count(&count).Error
	return math.Ceil(float64(count) / float64(filter.Size)), err
}

func (r *Repository) GetThemeById(userTGId int64, themeId uint) (types.ThemeModel, error) {
	var task types.ThemeModel
	query := r.Database.Model(&types.ThemeModel{})
	query = handleThemePreload(query)
	query = query.
		Where("theme_models.id=?", themeId).
		Where("user_models.tg_id=?", userTGId)
	err := query.First(&task).Error
	return task, err
}

// GetLastTheme - получить последнюю редактируемую задачу
func (r *Repository) GetLastTheme(userTGId int64) (types.ThemeModel, error) {
	var theme types.ThemeModel
	query := r.Database.Model(&types.ThemeModel{})
	query = handleThemePreload(query)
	query = query.
		Where("user_models.tg_id = ?", userTGId)
	query = query.Order("theme_models.updated_at desc")
	err := query.Last(&theme).Error
	return theme, err
}

// CreateTheme - создать тему
func (r *Repository) CreateTheme(theme types.ThemeModel) (types.ThemeModel, error) {
	return theme, r.Database.Create(&theme).Error
}

// UpdateTheme - обновить тему
func (r *Repository) UpdateTheme(theme types.ThemeModel) error {
	return r.Database.Model(types.ThemeModel{}).Where("id = ?", theme.ID).Updates(theme).Error
}

// DeleteTheme - удалить тему
func (r *Repository) DeleteTheme(id uint) error {
	return r.Database.Delete(&types.ThemeModel{}, id).Error
}

// handleTaskPreload - сформировать запрос подключения внешних таблиц к таблице тем и предварительно загрузить данные
func handleThemePreload(query *gorm.DB) *gorm.DB {
	query = query.
		Joins("left join user_models on user_models.id = theme_models.user_id")
	query = query.Preload("Tasks")
	query = query.Group("theme_models.id")
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
