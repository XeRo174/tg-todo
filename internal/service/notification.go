package service

import (
	"context"
	"fmt"
	"tg-todo/internal/types"
	"tg-todo/internal/utils"
	"time"
)

func (s *Service) StartNotification(ctx context.Context) {
	ticker := time.NewTicker(s.CheckInterval)
	defer ticker.Stop()
	fmt.Println("Запуск уведомлений")
	for {
		select {
		case <-ticker.C:
			s.check()
		case <-s.SignalChan:
			s.check()
		case <-ctx.Done():
			fmt.Println("Остановка")
			return
		}
	}
}

func (s *Service) check() {
	now := time.Now()
	fmt.Println("Запущена проверка работ")
	activeStatus := []types.TaskStatus{
		types.TaskStatusCreated,
		types.TaskStatusInWork,
	}
	for _, status := range activeStatus {
		tasks, err := s.Repository.GetTasks(types.TaskFilter{Status: int(status), SortQuery: types.SortQuery{Size: types.UnlimitedSize, Page: 1}})
		if err != nil {
			fmt.Println(fmt.Sprintf("ошибка получения задач со статусом %s: %v", status.String(), err))
			continue
		}
		for _, task := range tasks {
			loc, err := time.LoadLocation(task.User.TimeZone)
			if err != nil {
				message := fmt.Sprintf("ошибка загрузки зоны: %v", err)
				if _, err = s.Bot.SendMessage(task.User.ChatId, message, nil); err != nil {
					fmt.Println(fmt.Sprintf("ошибка отправки сообщения пользователю '%s': %v", message, err))
					continue
				}
				continue
			}
			nowInUserZone := now.In(loc)
			timeUntilDeadline := task.Deadline.Sub(nowInUserZone)

			var requiredLevelNotification int
			switch {
			case timeUntilDeadline <= 0:
				requiredLevelNotification = types.TaskLastCheckExpired
			case timeUntilDeadline <= 10*time.Minute:
				requiredLevelNotification = types.TaskLastCheck10M
			case timeUntilDeadline <= time.Hour:
				requiredLevelNotification = types.TaskLastCheck1H
			case timeUntilDeadline <= 8*time.Hour:
				requiredLevelNotification = types.TaskLastCheck8H
			case timeUntilDeadline <= 24*time.Hour:
				requiredLevelNotification = types.TaskLastCheck24H
			default:
				continue
			}
			if task.LastCheck >= requiredLevelNotification {
				continue
			}
			var message string
			switch requiredLevelNotification {
			case types.TaskLastCheckExpired:
				message = fmt.Sprintf("Уведомление. Задача %s просрочена, срок выполнения истек %s назад, статус %s, приоритет %s, темы %s",
					task.Name, formatDuration(-timeUntilDeadline), task.Status.String(), task.Priority.String(), utils.ThemeStroke(task.Themes))
			case types.TaskLastCheck10M:
				message = fmt.Sprintf("Уведомление. Задача %s. До дедлайна меньше 10 минут, осталось %s, статус %s, приоритет %s, темы %s",
					task.Name, formatDuration(timeUntilDeadline), task.Status.String(), task.Priority.String(), utils.ThemeStroke(task.Themes))
			case types.TaskLastCheck1H:
				message = fmt.Sprintf("Уведомление. Задача %s. До дедлайна меньше часа, осталось %s, статус %s, приоритет %s, темы %s",
					task.Name, formatDuration(timeUntilDeadline), task.Status.String(), task.Priority.String(), utils.ThemeStroke(task.Themes))
			case types.TaskLastCheck8H:
				message = fmt.Sprintf("Уведомление. Задача %s. До дедлайна меньше 8 часов, осталось %s, статус %s, приоритет %s, темы %s",
					task.Name, formatDuration(timeUntilDeadline), task.Status.String(), task.Priority.String(), utils.ThemeStroke(task.Themes))
			case types.TaskLastCheck24H:
				message = fmt.Sprintf("Уведомление. Задача %s. До дедлайна меньше 24 часов, осталось %s, статус %s, приоритет %s, темы %s",
					task.Name, formatDuration(timeUntilDeadline), task.Status.String(), task.Priority.String(), utils.ThemeStroke(task.Themes))
			}
			task.LastCheck = requiredLevelNotification
			if err = s.Repository.UpdateTask(task); err != nil {
				s.App.Logger.Error(fmt.Sprintf("%s. Обновление времени проверки задачи : %v", now.Format("02-15:04:05"), err))
			}
			if message == "" || task.User.ChatId == 0 {
				continue
			}
			if _, err = s.Bot.SendMessage(task.User.ChatId, message, nil); err != nil {
				s.App.Logger.Error(fmt.Sprintf("ошибка отправки сообщения пользователю : %v", err))
			}
		}
	}
}

func (s *Service) Close() {
	close(s.SignalChan)
	fmt.Println("канал закрыт")
}
func formatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dч %dм", hours, minutes)
	}
	return fmt.Sprintf("%dм", minutes)
}
