package logger

import (
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"os"
)

var (
	Info      *log.Logger
	Error     *log.Logger
	Debug     *log.Logger // Новый логгер для отладочных сообщений
	DebugMode = false     // Глобальная переменная для управления режимом отладки
)

func InitLoggers() {
	color.NoColor = false // Отключаем автоматическое определение поддержки цвета

	InfoColor := color.New(color.FgCyan).SprintFunc()
	ErrorColor := color.New(color.FgRed).SprintFunc()
	DebugColor := color.New(color.BgBlue).SprintFunc()

	// Логгеры с цветными сообщениями
	Info = log.New(os.Stdout, InfoColor("INFO: "), log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stdout, ErrorColor("ERROR: "), log.Ldate|log.Ltime|log.Lshortfile)
	// Проверяем, включен ли режим отладки

	if DebugMode {
		Debug = log.New(os.Stdout, DebugColor("DEBUG: "), log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		Debug = log.New(ioutil.Discard, "", 0) // Пустой вывод
	}
	Debug.Println("DEBUG MODE ON")
}
