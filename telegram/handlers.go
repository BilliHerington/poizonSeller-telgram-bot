package telegram

import (
	"crypto/rand"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/big"
	"net/url"
	"sailerBot/logger"
	"sailerBot/sheet"
	"strconv"
	"strings"
)

//------------------------------------------------------------------------------------------------MAIN------------------------------------------------------------------------------------

func HandleCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	logger.Debug.Println("HandleCallback running")
	mu.Lock()
	context, exists := userContexts[callbackQuery.Message.Chat.ID]
	if !exists {
		context = &UserContext{}
		userContexts[callbackQuery.Message.Chat.ID] = context
	}
	mu.Unlock()

	switch callbackQuery.Data {
	case "calculateShip":
		handleDeletePreviousMessage(bot, callbackQuery)
		handleGetProductPrice(bot, callbackQuery.Message, context)
	case "makeOrder":
		context.isMakingOrder = true
		handleDeletePreviousMessage(bot, callbackQuery)
		handleGetProductPrice(bot, callbackQuery.Message, context)
	case "exchangeRate":
		handleExchangeRate(bot, callbackQuery.Message)
	case "cart":
		handleDeletePreviousMessage(bot, callbackQuery)
		cart(bot, callbackQuery.Message)
	case "gotoPayment":
		gotoPayment(bot, callbackQuery.Message, context)
	case "sendVideo":
		sendVideo(bot, callbackQuery.Message)
	case "approvePayment":
		approvePayment(bot, callbackQuery.Message, context)
	case "changeData":
		changeData(bot, callbackQuery.Message, context)
	case "deleteProduct":
		deleteProductFromCart(bot, callbackQuery.Message, context)
	case "clearCart":
		clearCart(bot, callbackQuery.Message)
	//-----MENU------
	case "shoesMenu":
		productMenuCallback(bot, callbackQuery, context, "shoes")
	case "sneakersMenu":
		productMenuCallback(bot, callbackQuery, context, "sneakers")
	case "apparelMenu":
		productMenuCallback(bot, callbackQuery, context, "apparel")
	case "accessoriesMenu":
		productMenuCallback(bot, callbackQuery, context, "accessories")
	case "bagsMenu":
		productMenuCallback(bot, callbackQuery, context, "bags")
	case "toys&collectiblesMenu":
		productMenuCallback(bot, callbackQuery, context, "toys&collectibles")
	case "backMainMenu":
		context.isMakingOrder = false
		//context.
		handleDeletePreviousMessage(bot, callbackQuery)
		handleStartCommand(bot, callbackQuery.Message)
	}
}
func handleTextMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("handleTextMessage running")
	mu.Lock()
	context, exists := userContexts[message.Chat.ID]
	if !exists {
		context = &UserContext{}
		userContexts[message.Chat.ID] = context
	}
	mu.Unlock()
	logger.Debug.Printf("User send message: %s", message.Text)

	switch context.State {
	case StateWaitingForProductPrice:
		handleProductPrice(bot, message, context)
	case StateWaitingForProductURL:
		handleProductLink(bot, message, context)
	case StateWaitingForProductSize:
		handleSelectSize(bot, message, context)
	case StateWaitingForPhoneNumber:
		handlePhoneNumber(bot, message, context)
	case StateWaitingForFIO:
		handleFIO(bot, message, context)
	case StateWaitingForAddress:
		handleAddress(bot, message, context)
	case StateWaitingForDeletePosition:
		handleDeleteFromCart(bot, message, context)
	default:
		handleUnknownMessage(bot, message)
	}
}
func handleDeletePreviousMessage(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	logger.Debug.Println("handleDeletePreviousMessage running")
	deleteMessage := tgbotapi.NewDeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	_, err := bot.Request(deleteMessage)
	if err != nil {
		somethingWentWrong(bot, callbackQuery.Message)
		logger.Error.Println(err)
	}
}
func somethingWentWrong(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("somethingWentWrong running")
	msg := tgbotapi.NewMessage(message.Chat.ID, "Что-то пошло не так... Мы работаем над этим.")
	_, err := bot.Send(msg)
	if err != nil {
		logger.Error.Println(err)
	}
	gif := tgbotapi.NewAnimation(message.Chat.ID, tgbotapi.FilePath("config/telegram/images/cat.gif"))
	_, err = bot.Send(gif)
	if err != nil {
		logger.Error.Println(err)
	}
}
func handleStartCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("handleStartCommand running")
	msg := "Я poizonSellerBot, лучший инструмент для оформления заказа с сайта POIZON\n\n❗️ Если в процессе возникнут вопросы, свяжитесь с менеджером, он всегда готов помочь"
	photo := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FilePath("config/telegram/images/start.jpg"))
	photo.Caption = msg
	inlineKeyBoard := startMenuKeyboard()
	photo.ReplyMarkup = &inlineKeyBoard
	_, err := bot.Send(photo)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
}
func handleUnknownMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("handleUnknownMessage running")
	msg := tgbotapi.NewMessage(message.Chat.ID, "Я вас не помнимаю")
	_, err := bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
}

//--------------------------------------------------------------------------------------------MENU----------------------------------------------------------------------------------------

func makingOrder(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, context *UserContext) {
	logger.Debug.Println("makingOrder running")
	calculateShipCoastByProduct(callbackQuery)
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Отправте ссылку на товар 🔗")
	_, err := bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, callbackQuery.Message)
		logger.Error.Println(err)
	}
	context.State = StateWaitingForProductURL
}
func productMenuCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, context *UserContext, productCategory string) {
	logger.Debug.Println("productMenuCallback running")
	handleDeletePreviousMessage(bot, callbackQuery)
	setUserData(callbackQuery.Message.Chat.ID, "productCategory", productCategory)
	if context.isMakingOrder {
		makingOrder(bot, callbackQuery, context)
	} else {
		handleShipCoast(bot, callbackQuery, context)
	}
}
func sendVideo(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	video := tgbotapi.NewVideo(message.Chat.ID, tgbotapi.FilePath("config/telegram/images/howtoGetPrice.mp4"))
	_, err := bot.Send(video)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
}
func handleExchangeRate(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("handleRate running")
	text := "Почему курс юаня так высок?❓🇨🇳\nЦентральный банк РФ устанавливает официальный курс, который обычно отличается от реального, так как в банках и через посредников цена юаня в среднем на 3,5 рубля выше. Мы мгновенно конвертируем рубли в юани, стараясь предлагать лучшие условия по сравнению с конкурентами."
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Главное меню ☰", "backMainMenu")))
	_, err := bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
}
func categoryMenu(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("productsMenu running")
	msg := tgbotapi.NewMessage(message.Chat.ID, "Хорошо, а теперь выберите категорию в которую входит желаемый товар, это так же влияет на конечную цену")
	msg.ReplyMarkup = productMenuKeyboard()
	_, err := bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
}

//-------------------------------------------------------------------------------------------Products------------------------------------------------------------------------------------

func handleGetProductPrice(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("handleGetProductPrice running")
	messageText := "Укажите цену в юанях💸. Указывайте верную стоимость товара, иначе заказ будет отменён. Подробная инструкция показана на видео."
	msg := tgbotapi.NewMessage(message.Chat.ID, messageText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Получить видео🎬", "sendVideo")))
	_, err := bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
	context.State = StateWaitingForProductPrice
}

func handleProductPrice(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("handleProductPrice running")

	flPrice, err := strconv.ParseFloat(message.Text, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌Неверный формат цены")
		_, err = bot.Send(msg)
		if err != nil {
			somethingWentWrong(bot, message)
			logger.Error.Println(err)
		}
	} else {
		if flPrice < 0 {
			flPrice = -flPrice
		}
		redactedPrice := strconv.FormatFloat(flPrice, 'f', 0, 64)
		setUserData(message.Chat.ID, "productPriceRaw", redactedPrice)
		categoryMenu(bot, message)
	}
}
func handleProductLink(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("handleProductLink running")
	if isValidURL(message.Text) {
		setUserData(message.Chat.ID, "productURL", message.Text)

		logger.Debug.Println(getUserData(message.Chat.ID, "productURL"))

		msg := tgbotapi.NewMessage(message.Chat.ID, "🎨Отлично, теперь укажите размер и цвет. \nПример:\n11 US Men, Lucky Green")
		_, err := bot.Send(msg)
		if err != nil {
			somethingWentWrong(bot, message)
			logger.Error.Println(err)
		}

		context.State = StateWaitingForProductSize
	} else {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неверный формат ссылки❌. Повторно отправте ссылку на товар ввиде: \nhttps://www.poizon.com/product/ВАШ_ТОВАР")
		_, err := bot.Send(msg)
		if err != nil {
			somethingWentWrong(bot, message)
			logger.Error.Println(err)
		}
	}
}
func handleSelectSize(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("handleSelectSize running")
	setUserData(message.Chat.ID, "productSize", message.Text)
	cart(bot, message)
}

// --------------------------------------------------------------------------------USER-----------------------------------------------------------------------------------------------------
func handlePhoneNumber(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("handleUserPhoneNumber running")
	setUserData(message.Chat.ID, "userPhoneNumber", message.Text)

	userFIO, _ := getUserData(message.Chat.ID, "userFIO")
	if userFIO == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "👤ФИО\nВведите фамилию , имя и отчество(если имеется)\n\nНа указанное ФИО будет оформлена доставка")
		_, err := bot.Send(msg)
		if err != nil {
			somethingWentWrong(bot, message)
			logger.Error.Println(err)
		}
		context.State = StateWaitingForFIO
	}
}
func handleFIO(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("handleFIO running")
	setUserData(message.Chat.ID, "userFIO", message.Text)

	userAddress, _ := getUserData(message.Chat.ID, "userAddress")
	if userAddress == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "🏠 Данные адреса для отправки:\n\nУкажите адрес ближайшего пункта выдачи CDEK в формате Город, Улица, Дом")
		_, err := bot.Send(msg)
		if err != nil {
			somethingWentWrong(bot, message)
			logger.Error.Println(err)
		}
		context.State = StateWaitingForAddress
	}
}
func handleAddress(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("handleAddress running")
	setUserData(message.Chat.ID, "userAddress", message.Text)
	paymentMessage(bot, message, context)
}
func changeData(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("changeData running")
	setUserData(message.Chat.ID, "userPhoneNumber", "")
	setUserData(message.Chat.ID, "userFIO", "")
	setUserData(message.Chat.ID, "userAddress", "")
	msg := tgbotapi.NewMessage(message.Chat.ID, "📱Укажите номер телефона, для оформления заказа на доставку")
	_, err := bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
	context.State = StateWaitingForPhoneNumber
}

// --------------------------------------------------------------------------------CART-----------------------------------------------------------------------------------------------------
func cart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("cart running")
	addToCart(bot, message)
	readFromCart(bot, message)
}
func readFromCart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("readFromCart running")
	allDataFromCart, err := sheet.ReadCart(message.Chat.ID)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	allProductsText := ""
	allCartPrice := 0.0
	priceList := allDataFromCart["productPrice"]
	linkList := allDataFromCart["productLink"]
	sizeList := allDataFromCart["productSize"]
	if priceList != nil && linkList != nil && sizeList != nil {
		for idx := range priceList {
			allProductsText += fmt.Sprintf("№: %d\nЦена: %s\nСсылка: %s\nРазмер и Цвет: %s\n------------------------------\n", idx, priceList[idx], linkList[idx], sizeList[idx])
			floatValue, _ := strconv.ParseFloat(priceList[idx], 64)
			allCartPrice += floatValue
		}
		text := fmt.Sprintf(
			"Итоговая цена составит %0.f руб.\nВ эту сумму входит комиссия нашего сервиса, доставка по Китаю, доставка до РФ и страховка заказа.\n\n"+
				"🛍️ Товары в корзине:\n\n"+allProductsText,
			allCartPrice,
		)
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyMarkup = cartKeyboard()
		_, err = bot.Send(msg)
		if err != nil {
			somethingWentWrong(bot, message)
			logger.Error.Println(err)
		}
	} else {
		text := fmt.Sprintf("Ваша корзина пуста")
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyMarkup = emptyCartKeyboard()
		_, err = bot.Send(msg)
		if err != nil {
			somethingWentWrong(bot, message)
			logger.Error.Println(err)
		}
	}
}
func addToCart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("addToCart running")
	price, err := getUserData(message.Chat.ID, "productPrice")
	if err != nil {
		logger.Error.Println(err)
	}
	link, err := getUserData(message.Chat.ID, "productURL")
	if err != nil {
		logger.Error.Println(err)
	}
	size, err := getUserData(message.Chat.ID, "productSize")
	if err != nil {
		logger.Error.Println(err)
	}
	if price != "" && link != "" && size != "" {
		err = sheet.AddToCart(message.Chat.ID, price, link, size)
		if err != nil {
			logger.Error.Println(err)
			return
		}
	}
}
func deleteProductFromCart(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("deleteProductFromCart running")
	setUserData(message.Chat.ID, "productPrice", "")
	setUserData(message.Chat.ID, "productLink", "")
	setUserData(message.Chat.ID, "productSize", "")
	msg := tgbotapi.NewMessage(message.Chat.ID, "Напишите номер позиции одной цифрой")
	_, err := bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
	context.State = StateWaitingForDeletePosition
}
func handleDeleteFromCart(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	pos, err := strconv.Atoi(message.Text)
	text := ""
	if err != nil {
		text = "Неверный формат числа"
	} else {
		err = sheet.DeleteFromCart(message.Chat.ID, pos)
		if err != nil {
			text = fmt.Sprintf("Товар № %d не найден в корзине", pos)
		}
		text = fmt.Sprintf("Товар № %d успешно удален из корзины", pos)
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Моя корзина🛒", "cart")))
	_, err = bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
}

func clearCart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	logger.Debug.Println("clearCart running")
	setUserData(message.Chat.ID, "productPrice", "")
	setUserData(message.Chat.ID, "productLink", "")
	setUserData(message.Chat.ID, "productSize", "")
	err := sheet.ClearCart(message.Chat.ID)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, "Ваша корзина успешно очищена")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Главное меню ☰", "backMainMenu")))
	_, err = bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
}

// -------------------------------------------------------------------------------PAYMENT-----------------------------------------------------------------------------------------------------
func gotoPayment(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("gotoPayment running")
	userPhoneNumber, _ := getUserData(message.Chat.ID, "userPhoneNumber")
	if userPhoneNumber == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "📱Укажите номер телефона, для оформления заказа на доставку")
		_, err := bot.Send(msg)
		if err != nil {
			somethingWentWrong(bot, message)
			logger.Error.Println(err)
		}
		context.State = StateWaitingForPhoneNumber
	} else {
		paymentMessage(bot, message, context)
	}
}
func paymentMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	//paymentPath := "config/payment/payment.env"
	//paymentData := loadPaymentENV(paymentPath)

	token, err := generateOrderKey("*", 4) // Генерируем ключ длиной 4 символа для каждой части
	if err != nil {
		logger.Error.Println("Error:", err)
		somethingWentWrong(bot, message)
	}

	name, err := getUserData(message.Chat.ID, "userFIO")
	if err != nil {
		logger.Error.Println("Error:", err)
	}

	number, err := getUserData(message.Chat.ID, "userPhoneNumber")
	if err != nil {
		logger.Error.Println("Error:", err)
	}

	address, err := getUserData(message.Chat.ID, "userAddress")
	if err != nil {
		logger.Error.Println("Error:", err)
	}
	payData := loadPaymentDataJSON()
	text := fmt.Sprintf("Данные для отправки:\nФИО: %s\nАдресс: %s\nНомер телефона: %s\n\n📦Доставка по РФ оплачивается отдельно напрямую СДЭКу (≈500₽)\n\n"+
		"⚠️Выкуп товара происходит в течении 24 часов после оплаты. \nВ случае если при выкупе заказа изменится цена , с вами свяжется менеджер\n\n"+
		"✨Это ваш уникальный номер заказа, сохраните его:\n%s\n\nДанные для оплаты:\n\n"+
		"💳СБЕР- %s\n\n🎫Тинькофф %s\n\n"+
		"📝Произведя оплату Вы соглашаетесь с корректностью данных и характеристик указанного товара"+
		"\n\nОплатите и нажмите кнопку Подтвердить оплату ✅", name, address, number, token, payData.Sber, payData.Tinkoff)
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Подтвердить оплату ✅", "approvePayment")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Изменить данные 🔑", "changeData")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Главное меню ☰", "backMainMenu")))
	_, err = bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, message)
		logger.Error.Println(err)
	}
	context.token = token
}
func approvePayment(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	logger.Debug.Println("approvePayment running")
	alertMsg := tgbotapi.NewMessage(message.Chat.ID, "Ваш заказ формируется⌛\n")
	_, err := bot.Send(alertMsg)
	if err != nil {
		somethingWentWrong(bot, message)
	}

	addProducts(bot, message, context)
	addOrder(bot, message, context)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Заказ успешно оформлен✅\nПозже с вами свяжется наш менеджер")
	_, err = bot.Send(msg)
	if err != nil {
		logger.Error.Println(err)
		somethingWentWrong(bot, message)
	}
}
func addProducts(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	allDataFromCart, readCartError := sheet.ReadCart(message.Chat.ID)
	if readCartError != nil {
		logger.Error.Println(readCartError)
		somethingWentWrong(bot, message)
	}
	priceList := allDataFromCart["productPrice"]
	linkList := allDataFromCart["productLink"]
	sizeList := allDataFromCart["productSize"]

	for idx, value := range priceList {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			logger.Error.Println(err)
			somethingWentWrong(bot, message)
		}
		err = sheet.AddProductsByToken(context.token, linkList[idx], sizeList[idx], intValue)
		if err != nil {
			logger.Error.Println(err)
			somethingWentWrong(bot, message)
		}
	}
}
func addOrder(bot *tgbotapi.BotAPI, message *tgbotapi.Message, context *UserContext) {
	name, err := getUserData(message.Chat.ID, "userFIO")
	if err != nil {
		logger.Error.Println(err)
		somethingWentWrong(bot, message)
	}
	address, err := getUserData(message.Chat.ID, "userAddress")
	if err != nil {
		logger.Error.Println(err)
		somethingWentWrong(bot, message)
	}
	phone, err := getUserData(message.Chat.ID, "userPhoneNumber")
	if err != nil {
		logger.Error.Println(err)
		somethingWentWrong(bot, message)
	}
	err = sheet.AddOrder(message.Chat.ID, context.token, name, address, phone, "PaymentNotApproved")
	if err != nil {
		logger.Error.Println(err)
		somethingWentWrong(bot, message)
	}
}

// --------------------------------------------------------------------------------OTHER-----------------------------------------------------------------------------------------------------
func isValidURL(rawURL string) bool {
	logger.Debug.Println("isValidURL running")
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme != "https" {
		return false
	}
	if parsedURL.Host != "www.poizon.com" {
		return false
	}
	if !strings.HasPrefix(parsedURL.Path, "/product/") {
		return false
	}
	return true
}
func calculateShipCoastByProduct(callbackQuery *tgbotapi.CallbackQuery) {
	logger.Debug.Println("calculateShipCoastByProduct running")

	productPriceRaw, err := getUserData(callbackQuery.Message.Chat.ID, "productPriceRaw")
	if err != nil {
		logger.Error.Println(err)
	}
	productPriceFloat, err := strconv.ParseFloat(productPriceRaw, 64)

	if err != nil {
		logger.Error.Println(err)
	}
	productCategory, err := getUserData(callbackQuery.Message.Chat.ID, "productCategory")
	if err != nil {
		logger.Error.Println(err)
	}
	resultPrice := CnyExRate(productPriceFloat) + getCoast(productCategory)

	setUserData(callbackQuery.Message.Chat.ID, "productPrice", strconv.FormatFloat(resultPrice, 'f', 0, 64))
}
func handleShipCoast(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, context *UserContext) {
	logger.Debug.Println("handleShipCoast running")
	calculateShipCoastByProduct(callbackQuery)
	resultPrice, err := getUserData(callbackQuery.Message.Chat.ID, "productPrice")
	if err != nil {
		logger.Error.Println(err)
	}
	msgText := fmt.Sprintf("Расчётная стоимость заказа %s руб.", resultPrice)
	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, msgText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Главное меню ☰", "backMainMenu")))

	_, err = bot.Send(msg)
	if err != nil {
		somethingWentWrong(bot, callbackQuery.Message)
		logger.Error.Println(err)
	}
}

func generateOrderKey(prefix string, length int) (string, error) {
	part1, err := generateRandomString(length)
	if err != nil {
		return "", err
	}
	// Генерация второй части ключа
	part2, err := generateRandomString(length)
	if err != nil {
		return "", err
	}
	// Формирование окончательного ключа

	orderKey := fmt.Sprintf("%s%s-%s", prefix, part1, part2)
	return orderKey, nil
}
func generateRandomString(length int) (string, error) {
	letterBytes := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		if err != nil {
			return "", err
		}
		result += string(letterBytes[index.Int64()])
	}
	return result, nil
}
