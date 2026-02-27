package bootstrap

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"log/slog"
	"os"
	"strings"
	"tg-todo/internal/types"
)

type Environment struct {
	UpdateMode       string `env:"UPDATE_MODE"`
	Token            string `env:"TG_TOKEN"`
	LogFile          string `env:"LOG_FILE"`
	DatabaseType     string `env:"DATABASE_TYPE"`     // Тип базы данных
	DatabaseFile     string `env:"DATABASE_FILE"`     // Файл базы данных
	DatabaseHost     string `env:"DATABASE_HOST"`     // Хост базы данных
	DatabasePort     string `env:"DATABASE_PORT"`     // Порт базы данных
	DatabaseName     string `env:"DATABASE_NAME"`     // Имя базы данных
	DatabaseUser     string `env:"DATABASE_USER"`     // Пользователь базы данных
	DatabasePassword string `env:"DATABASE_PASSWORD"` // Пароль базы данных
}

type Application struct {
	Environment Environment
	Logger      *slog.Logger
	Database    *gorm.DB
}

func App() Application {
	env := setupEnv()
	logger := setupLogger(env.LogFile)
	conn := connectToDB(env.DatabaseFile)
	app := Application{
		Environment: env,
		Logger:      logger,
		Database:    conn,
	}
	return app
}

func setupEnv() Environment {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")
	env := Environment{}
	ctx := context.Background()
	if err := envconfig.Process(ctx, &env); err != nil {
		logrus.Fatalf("Ошибка загрузки параметров env: %v", err)
	}
	return env
}

func setupLogger(logFile string) *slog.Logger {
	logWriter := &lumberjack.Logger{
		Filename:  logFile,
		Compress:  true,
		MaxSize:   1,
		LocalTime: true,
	}
	output := io.MultiWriter(logWriter, os.Stdout)
	logger := slog.New(slog.NewTextHandler(output, nil))
	return logger

}

func setupLogrus(logFile string) *logrus.Logger {
	logger := logrus.New()
	TextFormatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05 02-01-2006",
		DisableQuote:    true,
		PadLevelText:    true,
	}
	logWriter := &lumberjack.Logger{
		Filename:  logFile,
		Compress:  true,
		MaxSize:   1,
		LocalTime: true,
	}
	output := io.MultiWriter(logWriter, os.Stdout)
	logger.SetFormatter(TextFormatter)
	logger.SetOutput(output)
	logger.SetLevel(logrus.InfoLevel)
	logger.Info(strings.Repeat("-", 10))
	logger.Info("Сервер запущен")
	logger.Info(strings.Repeat("-", 10))
	return logger
}

func connectToDB(databaseFile string) *gorm.DB {
	var conn *gorm.DB
	var err error
	if conn, err = gorm.Open(sqlite.Open(databaseFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}); err != nil {
		panic(err)
	}
	conn.AutoMigrate(&types.UserModel{}, &types.ThemeModel{}, &types.TaskModel{})
	return conn
}
