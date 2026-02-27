package types

import (
	"fmt"
	"strings"
)

const (
	TaskPriorityNone = 1 + iota
	TaskPriorityNormal
	TaskPriorityHigh
)

const (
	TaskStatusCreated = 1 + iota
	TaskStatusInWork
	TaskStatusDone
	TaskStatusFailed
	TaskStatusDropped
)

type Status struct {
	Name  string
	Value int
}

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

var Statuses = []Status{
	{
		Name:  "Создана",
		Value: TaskStatusCreated,
	},
	{
		Name:  "В работе",
		Value: TaskStatusInWork,
	},
	{
		Name:  "Завершена",
		Value: TaskStatusDone,
	},
	{
		Name:  "Провалена",
		Value: TaskStatusFailed,
	},
	{
		Name:  "Заброшена",
		Value: TaskStatusDropped,
	},
}

func GetPriorityByValue(value int) Priority {
	for _, priority := range Priorities {
		if priority.Value == value {
			return priority
		}
	}
	return Priority{
		Name:  "Проблема определения",
		Value: -1,
	}
}

func GetPriorityByCallbackData(cbData string) Priority {
	cbData = strings.Replace(cbData, "task_priority:", "", 1)
	for _, priority := range Priorities {
		if cbData == fmt.Sprintf("%d", priority.Value) {
			return priority
		}
	}
	return Priority{
		Name:  "Проблема определения",
		Value: -1,
	}
}

func GetStatusByValue(value int) Status {
	for _, status := range Statuses {
		if status.Value == value {
			return status
		}
	}
	return Status{
		Name:  "Проблема определения",
		Value: -1,
	}
}

func GetStatusByCallbackData(cbData string) Status {
	cbData = strings.Replace(cbData, "task_status:", "", 1)
	for _, status := range Statuses {
		if cbData == fmt.Sprintf("%d", status.Value) {
			return status
		}
	}
	return Status{
		Name:  "Проблема определения",
		Value: -1,
	}
}
