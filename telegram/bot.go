package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"math"
	"os"
	"sailerBot/logger"
	"strconv"
	"sync"
)

type ConfigBot struct {
	APIkey string `env:"APIKEY"`
}
type PaymentData struct {
	Tinkoff string `env:"TINKOFF"`
	Sber    string `env:"SBER"`
	Other   string `env:"OTHER"`
}

func LoadConfig(filename string) string {
	err := godotenv.Load(filename)
	if err != nil {
		logger.Error.Fatalf("Error loading .env file %v", err)
	}
	config := ConfigBot{}

	err = envconfig.Process("", &config)
	if err != nil {
		logger.Error.Fatalf("Error processing .env file %v", err)
	}
	return config.APIkey
}
func loadPaymentENV(filename string) PaymentData {
	err := godotenv.Load(filename)
	if err != nil {
		logger.Error.Printf("Error loading .env file %v", err)
	}
	config := PaymentData{}
	err = envconfig.Process("", &config)
	if err != nil {
		logger.Error.Printf("Error processing .env file %v", err)
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

//------------------------------------------------------------------

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
	err := godotenv.Load("config/telegram/shippingPrices/shippingPrices.env") // Укажите путь к вашему .env файлу
	if err != nil {
		logger.Error.Println("Error loading .env file: %v", err)
	}

	// Пример заполнения карты значениями из env
	ShippingProducts["shoes"] = getEnvFloat("SHOES")
	ShippingProducts["sneakers"] = getEnvFloat("SNEAKERS")
	ShippingProducts["apparel"] = getEnvFloat("APPAREL")
	ShippingProducts["bags"] = getEnvFloat("BAGS")
	ShippingProducts["accessories"] = getEnvFloat("ACCESSORIES")
	ShippingProducts["toys&collectibles"] = getEnvFloat("TOYS_COLLECTIBLES")
}

// Вспомогательная функция для получения значения из env как float64
func getEnvFloat(key string) float64 {
	valStr := os.Getenv(key)
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		logger.Error.Printf("Error parsing env variable %s: %v", key, err)
		return 0
	}
	return val
}
func getCoast(name string) float64 {
	if value, exist := ShippingProducts[name]; exist {
		return value
	}
	logger.Error.Println("Shipping coast product not found")
	return 0
}
