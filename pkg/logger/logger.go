package logger

import (
	"log"
	"os"
)

var (
	projectLogger  *log.Logger
	telegramLogger *log.Logger
)

// Init инициализирует логгеры для проекта и телеги
func Init(projectLogPath, telegramLogPath string) error {

	projectFile, err := os.OpenFile(projectLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	telegramFile, err := os.OpenFile(telegramLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	projectLogger = log.New(projectFile, "PROJECT ", log.Ldate|log.Ltime|log.Lshortfile)
	telegramLogger = log.New(telegramFile, "TELEGRAM ", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func Info(msg string, args ...interface{}) {
	if projectLogger != nil {
		projectLogger.Printf("INFO: "+msg, args...)
	} else {
		log.Printf("INFO: "+msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	if projectLogger != nil {
		projectLogger.Printf("ERROR: "+msg, args...)
	} else {
		log.Printf("ERROR: "+msg, args...)
	}
}

func TelegramInfo(msg string, args ...interface{}) {
	if telegramLogger != nil {
		telegramLogger.Printf("INFO: "+msg, args...)
	} else {
		log.Printf("TELEGRAM INFO: "+msg, args...)
	}
}

func TelegramError(msg string, args ...interface{}) {
	if telegramLogger != nil {
		telegramLogger.Printf("ERROR: "+msg, args...)
	} else {
		log.Printf("TELEGRAM ERROR: "+msg, args...)
	}
}
