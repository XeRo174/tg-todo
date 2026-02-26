package types

import (
	"gorm.io/gorm"
	"strings"
	"time"
)

const (
	TaskPriorityNone = iota
	TaskPriorityNormal
	TaskPriorityHigh
)

type Priority struct {
	Name  string
	Value int
}

var Priorities = []Priority{
	{
		Name:  "Без приоритета",
		Value: TaskPriorityNone,
	},
	{
		Name:  "Обычный",
		Value: TaskPriorityNormal,
	},
	{
		Name:  "Высокий приоритет",
		Value: TaskPriorityHigh,
	},
}

func GetPriorityByCallbackData(cbData string) Priority {
	cbData = strings.Replace(cbData, "task_priority:", "", 1)
	for _, priority := range Priorities {
		if cbData == string(rune(priority.Value)) {
			return priority
		}
	}
	return Priority{}
}

type UserModel struct {
	gorm.Model
	TGId     int64
	Username string
	TimeZone string
}

type TaskThemeModel struct {
	gorm.Model
	Name        string
	User        UserModel
	UserId      uint
	TaskModelID uint
}

type TaskModel struct {
	gorm.Model
	Name     string
	Status   int
	Priority int
	User     UserModel
	UserId   uint
	Themes   []TaskThemeModel `gorm:"foreignKey:TaskModelID"`
	Deadline time.Time
	Editable bool
}
