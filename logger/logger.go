package logger

import (
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"log"
	"os"
)

var (
	Info  *log.Logger
	Error *log.Logger
	Debug *log.Logger // Новый логгер для отладочных сообщений
	// Глобальная переменная для управления режимом отладки
)

type DebugEnv struct {
	DebugMode bool `env:"DEBUG_MODE"`
}

func loadDebugOptionEnv() bool {
	err := godotenv.Load("config/logSettings/debugMode.env")
	if err != nil {
		log.Fatal("Error loading debugMode.env file")
	}
	debug := DebugEnv{}
	err = envconfig.Process("", &debug)
	if err != nil {
		log.Fatal("Error processing debug debugMode.env file")
	}
	return debug.DebugMode
}
func InitLoggers() {
	color.NoColor = false // Отключаем автоматическое определение поддержки цвета

	InfoColor := color.New(color.FgCyan).SprintFunc()
	ErrorColor := color.New(color.FgRed).SprintFunc()
	DebugColor := color.New(color.BgBlue).SprintFunc()

	// Логгеры с цветными сообщениями
	Info = log.New(os.Stdout, InfoColor("INFO: "), log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stdout, ErrorColor("ERROR: "), log.Ldate|log.Ltime|log.Lshortfile)
	// Проверяем, включен ли режим отладки
	DebugMode := loadDebugOptionEnv()
	if DebugMode {
		Debug = log.New(os.Stdout, DebugColor("DEBUG: "), log.Ldate|log.Ltime|log.Lshortfile)
		Debug.Println("DEBUG MODE ON")
	} else {
		Debug = log.New(ioutil.Discard, "", 0) // Пустой вывод
	}
	
}
