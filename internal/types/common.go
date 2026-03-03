package types

import (
	"fmt"
	"strings"
)

type TaskPriority int
type TaskStatus int

const (
	TaskPriorityNone TaskPriority = 1 + iota
	TaskPriorityNormal
	TaskPriorityHigh
)

const (
	TaskStatusDraft TaskStatus = 1 + iota
	TaskStatusCreated
	TaskStatusInWork
	TaskStatusDone
	TaskStatusFailed
	TaskStatusDropped
)

func (p TaskPriority) String() string {
	switch p {
	case TaskPriorityNone:
		return "Без приоритета"
	case TaskPriorityNormal:
		return "Обычный"
	case TaskPriorityHigh:
		return "Высокий"
	default:
		return "Не указан"
	}
}

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusDraft:
		return "Черновик"
	case TaskStatusCreated:
		return "Создана"
	case TaskStatusInWork:
		return "В работе"
	case TaskStatusDone:
		return "Выполнена"
	case TaskStatusFailed:
		return "Провалена"
	case TaskStatusDropped:
		return "Брошена"
	default:
		return "Не указан"
	}
}

func AllPriorities() []TaskPriority {
	return []TaskPriority{TaskPriorityNone, TaskPriorityNormal, TaskPriorityHigh}
}

func AllStatuses() []TaskStatus {
	return []TaskStatus{TaskStatusCreated, TaskStatusInWork, TaskStatusDone, TaskStatusFailed, TaskStatusDropped}
}

func ParsePriority(s string) (TaskPriority, bool) {
	s = strings.TrimPrefix(s, CallbackTaskPrioritySet)
	for _, p := range AllPriorities() {
		if fmt.Sprintf("%d", p) == s {
			return p, true
		}
	}
	return -1, false
}

func ParseStatus(s string) (TaskStatus, bool) {
	s = strings.TrimPrefix(s, CallbackTaskStatusSet)
	for _, p := range AllStatuses() {
		if fmt.Sprintf("%d", p) == s {
			return p, true
		}
	}
	return -1, false
}
