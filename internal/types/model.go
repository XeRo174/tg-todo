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
	Name   string `gorm:"uniqueIndex:idx_theme_user_name,priority:2,WHERE:deleted_at IS NULL"`
	User   UserModel
	UserId uint        `gorm:"uniqueIndex:idx_theme_user_name,priority:1,WHERE:deleted_at IS NULL"`
	Tasks  []TaskModel `gorm:"many2many:task_themes;"`
}

type TaskModel struct {
	gorm.Model
	Name     string
	Status   TaskStatus
	Priority TaskPriority
	User     UserModel
	UserId   uint
	Themes   []ThemeModel `gorm:"many2many:task_themes;"`
	Deadline time.Time
	Messages []MessageRegisterModel `gorm:"foreignkey:TaskId"`
}

type MessageRegisterModel struct {
	gorm.Model
	BotMessageId int64
	Task         TaskModel
	TaskId       uint
	Operation    string
}
