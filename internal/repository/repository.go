package repository

import (
	"gorm.io/gorm"
)

type Repository struct {
	Database *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Database: db,
	}
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

// wrapCondition - проверяет необходимость точного поиска и формирует условие
func wrapCondition(likeSearch bool, value string) (string, string) {
	if likeSearch && value != "" {
		return "LIKE ?", "%" + value + "%"
	}
	return "=?", value
}

// setPagination - устанавливает лимит выводимых данных
func setPagination(query *gorm.DB, size int, page uint) *gorm.DB {
	query = query.Limit(size).Limit(size)
	if page > 1 {
		query = query.Offset(int(page-1) * size)
	}
	return query
}
