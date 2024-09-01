package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sailerBot/logger"
	"sailerBot/telegram"
	"sync"
)

//TIP To run your code, right-click the code and select <b>Run</b>. Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.

// TODO config
// TODO logger
// TODO тгбот
// TODO poizonAPI
// TODO database
func main() {

	logger.InitLoggers()

	var wg sync.WaitGroup
	wg.Add(1)
	configPath := "config/telegram/bot.env"
	configBot := telegram.LoadConfig(configPath)
	botAPI, err := tgbotapi.NewBotAPI(configBot)
	if err != nil {
		logger.Error.Fatalln(err)
	}

	botAPI.Debug = false
	logger.Info.Printf("Authorized on account %s", botAPI.Self.UserName)
	telegram.InitShippingProducts()
	go func() {
		defer wg.Done()
		telegram.RunBot(botAPI)
	}()
	wg.Wait()
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
