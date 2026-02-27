package types

import (
	"gorm.io/gorm"
	"time"
)

type UserModel struct {
	gorm.Model
	TGId     int64
	Username string
	TimeZone string
}

type ThemeModel struct {
	gorm.Model
	Name   string
	User   UserModel
	UserId uint
	Tasks  []TaskModel `gorm:"many2many:task_themes;"`
}

type TaskModel struct {
	gorm.Model
	Name     string
	Status   int
	Priority int
	User     UserModel
	UserId   uint
	Themes   []ThemeModel `gorm:"many2many:task_themes;"`
	Deadline time.Time
	Editable bool
}
