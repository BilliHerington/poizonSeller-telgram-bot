package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"os"
	"sailerBot/logger"
	"sailerBot/telegram"
)

func main() {

	logger.InitLoggers()
	telegram.InitShippingProducts()

	if err := godotenv.Load("config/.env"); err != nil {
		logger.Error.Println("Error loading .env file", err)
		return
	}
	botAPI, err := tgbotapi.NewBotAPI(os.Getenv("TG_API_KEY"))
	if err != nil {
		logger.Error.Fatalln(err)
	}
	logger.Info.Printf("Authorized on account %s", botAPI.Self.UserName)

	botAPI.Debug = false
	telegram.RunBot(botAPI)
}
