package manager

import (
	"fmt"
	"tg-todo/internal/types"
	"time"
)

type Manager struct {
	Repository interface {
		GetTasks(filter types.TaskFilter) ([]types.TaskModel, error)
	}
	CheckInterval time.Duration
	stopChan      chan struct{}
}

func NewManager(repo interface {
	GetTasks(filter types.TaskFilter) ([]types.TaskModel, error)
}, checkInterval time.Duration) *Manager {
	return &Manager{
		Repository:    repo,
		CheckInterval: checkInterval,
		stopChan:      make(chan struct{}),
	}
}

func (m *Manager) Start() {
	go m.run()
}

func (m *Manager) run() {
	ticker := time.NewTicker(m.CheckInterval)
	defer ticker.Stop()

	fmt.Println("Менеджер запущен")
	for {
		select {
		case <-ticker.C:
			m.check()
		case <-m.stopChan:
			fmt.Println("Менеджер остановлен")
			return
		}
	}
}

func (m *Manager) check() {
	now := time.Now()
	fmt.Println("49")
	activeStatus := []types.TaskStatus{
		types.TaskStatusCreated,
		types.TaskStatusInWork,
	}
	for _, status := range activeStatus {
		tasks, err := m.Repository.GetTasks(types.TaskFilter{Status: int(status), SortQuery: types.SortQuery{Size: types.UnlimitedSize, Page: 1}})
		if err != nil {
			fmt.Println(fmt.Sprintf("Ошибка получения задач со статусом %s: %w", status.String(), err))
		} else {
			fmt.Println("задачи получены")
		}
		for _, task := range tasks {
			loc, err := time.LoadLocation(task.User.TimeZone)
			if err != nil {
				fmt.Println("ошибка загрузки зоны: %w", err)
				//todo пометка для пользователя, что время отправляется не по его времени?.
			} else {
				fmt.Println("часовая зона загружена")
			}
			nowInUserZone := now.In(loc)
			fmt.Println(fmt.Sprintf("nowInUserZone %s", nowInUserZone))
			timeUntilDeadline := task.Deadline.Sub(nowInUserZone)
			fmt.Println(fmt.Sprintf("timeUntilDeadline %s", timeUntilDeadline))
			switch {
			case timeUntilDeadline <= 0:
				fmt.Println(fmt.Sprintf("Задача %s просрочена, срок был %s, опоздание %s, время по часовой зоне %s", task.Name, task.Deadline, timeUntilDeadline, task.User.TimeZone))
			case timeUntilDeadline <= time.Minute*10:
				fmt.Println(fmt.Sprintf("Задача %s, времени осталось меньше 10 минут, срок был %s, осталось %s, время по часовой зоне %s", task.Name, task.Deadline, timeUntilDeadline, task.User.TimeZone))
			case timeUntilDeadline <= time.Hour:
				fmt.Println(fmt.Sprintf("Задача %s, времени осталось меньше часа, срок был %s, осталось %s, время по часовой зоне %s", task.Name, task.Deadline, timeUntilDeadline, task.User.TimeZone))
			case timeUntilDeadline <= time.Hour*8:
				fmt.Println(fmt.Sprintf("Задача %s, времени осталось меньше 8 часов, срок был %s, осталось %s, время по часовой зоне %s", task.Name, task.Deadline, timeUntilDeadline, task.User.TimeZone))
			case timeUntilDeadline <= time.Hour*24:
				fmt.Println(fmt.Sprintf("Задача %s, времени осталось меньше 24 часов, срок был %s, осталось %s, время по часовой зоне %s", task.Name, task.Deadline, timeUntilDeadline, task.User.TimeZone))
			default:
				fmt.Println(fmt.Sprintf("время еще есть для задачи %s, у неё срок %s", task.Name, task.Deadline))
			}

			if task.Deadline == time.Now().In(task.Deadline.Location()) {
				fmt.Println("Пришло время выполнить задачу", task.Name)

			} else if task.Deadline.Before(time.Now().In(task.Deadline.Location())) {
				fmt.Println("Задача просрочена ", task.Name)
			} else {
				fmt.Println("Еще не время для задачи ", task.Name)
			}
		}
	}
}

func (m *Manager) Close() {
	close(m.stopChan)
	fmt.Println("канал закрыт")
}
