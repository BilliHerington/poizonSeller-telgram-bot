package telegram

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math"
	"os"
	"sailerBot/logger"
	"sync"
)

type ConfigBot struct {
	APIkey string `env:"APIKEY"`
}
type PaymentData struct {
	Tinkoff string `json:"tinkoff"`
	Sber    string `json:"sber"`
	Other   string `json:"other"`
}

//	func LoadConfig(filename string) string {
//		err := godotenv.Load(filename)
//		if err != nil {
//			logger.Error.Fatalf("Error loading .env file %v", err)
//		}
//		config := ConfigBot{}
//
//		err = envconfig.Process("", &config)
//		if err != nil {
//			logger.Error.Fatalf("Error processing .env file %v", err)
//		}
//		return config.APIkey
//	}
func loadPaymentDataJSON() PaymentData {
	file, err := os.ReadFile("config/telegram/payment/payment_data.json")
	if err != nil {
		logger.Error.Printf("Error reading paymnet json file %v", err)
		return PaymentData{}
	}
	config := PaymentData{}

	err = json.Unmarshal(file, &config)
	if err != nil {
		logger.Error.Printf("Error processing .JSON file %v", err)
		return PaymentData{}
	}
	//logger.Debug.Printf("Config info: tinl: %s, sber: %s, other: %s", config.Tinkoff, config.Sber, config.Other)
	return config
}
func RunBot(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.CallbackQuery != nil {
			// Подтверждаем получение callback-запроса
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			if _, err := bot.Request(callback); err != nil {
				logger.Error.Println("Error sending callback response:", err)
			}
			logger.Debug.Printf("callback query: %v", update.CallbackQuery)
			HandleCallback(bot, update.CallbackQuery)
		}

		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				handleStartCommand(bot, update.Message)
			}
		} else {
			handleTextMessage(bot, update.Message)
		}
	}
}

// ------------Реализация хранения данных о заказе в памяти-----------

// In-memory map для хранения данных о продуктах пользователей

var (
	userDataMap = make(map[int64]map[string]string)
	mu          sync.RWMutex
)

// Функция для установки данных о продукте
func setUserData(userID int64, userType string, value string) {
	mu.Lock()
	defer mu.Unlock()

	// Проверяем, существует ли запись для пользователя, если нет — создаём новую
	if _, exists := userDataMap[userID]; !exists {
		userDataMap[userID] = make(map[string]string)
	}

	// Сохраняем данные
	userDataMap[userID][userType] = value
}

// Функция для получения данных о продукте
func getUserData(userID int64, userType string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	// Проверяем, существует ли запись для пользователя
	userData, exists := userDataMap[userID]
	if !exists {
		return "", nil // Ключ не найден
	}

	// Проверяем, существует ли ключ для данного пользователя
	value, keyExists := userData[userType]
	if !keyExists {
		return "", nil // Ключ не найден
	}

	return value, nil
}

type UserState int

const (
	StateWaitingForProductPrice UserState = iota
	StateWaitingForProductURL
	StateWaitingForProductSize
	StateWaitingForPhoneNumber
	StateWaitingForFIO
	StateWaitingForAddress
	StateWaitingForDeletePosition
	// Добавьте другие состояния по мере необходимости
)

type UserContext struct {
	State         UserState
	isMakingOrder bool
	token         string
}

var userContexts = make(map[int64]*UserContext)

func CnyExRate(price float64) float64 {
	return math.Ceil(price * 12.5)
}

var ShippingProducts = map[string]float64{}

func InitShippingProducts() {
	file, err := os.ReadFile("config/telegram/shippingPrices/shippingPrices.json") // Укажите путь к вашему файлу
	if err != nil {
		logger.Error.Println("Error loading json file: %v", err)
		return
	}
	// Парсинг JSON в карту
	var shippingData map[string]float64
	err = json.Unmarshal(file, &shippingData)
	if err != nil {
		logger.Error.Println("Error parsing JSON file: %v", err)
		return
	}
	// Заполнение карты ShippingProducts из распарсенных данных
	for product, price := range shippingData {
		ShippingProducts[product] = price
	}

}

func getCoast(name string) float64 {
	if value, exist := ShippingProducts[name]; exist {
		return value
	}
	logger.Error.Println("Shipping coast product not found")
	return 0
}
