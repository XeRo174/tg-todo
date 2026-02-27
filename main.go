package main

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"tg-todo/internal/bootstrap"
	"tg-todo/internal/repository"
	"tg-todo/internal/service"
)

func main() {

	app := bootstrap.App()
	r := repository.NewRepository(app.Database)
	bot, err := gotgbot.NewBot(app.Environment.Token, nil)
	if err != nil {
		app.Logger.Error(fmt.Sprintf("Ошибка создания бота, err: %v", err))
		return
	}
	s := service.NewService(r, bot, app)

	s.Start()
}
