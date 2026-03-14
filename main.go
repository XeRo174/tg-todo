package main

import (
	"context"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"os"
	"os/signal"
	"syscall"
	"tg-todo/internal/bootstrap"
	"tg-todo/internal/repository"
	"tg-todo/internal/service"
	"time"
)

func main() {

	app := bootstrap.App()
	r := repository.NewRepository(app.Database)
	ctx, cancel := context.WithCancel(context.Background())
	bot, err := gotgbot.NewBot(app.Environment.Token, nil)
	if err != nil {
		app.Logger.Error(fmt.Sprintf("Ошибка создания бота, err: %v", err))
		return
	}
	s := service.NewService(r, bot, app, time.Minute*5)
	go s.StartNotification(ctx)

	s.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Получен сигнал завершения, начинаем graceful shutdown")
	s.Close()
	cancel()
}
