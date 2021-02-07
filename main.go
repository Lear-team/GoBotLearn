package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	sqlapi "GoBotPigeon/app/sqlapi"
	config "GoBotPigeon/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/kyokomi/emoji"
)

var configParam config.Configuration

var command string = ""

func main() {

	configParam, err := InitConfig()
	if err != nil {
		log.Panic(err)
	}

	err = sqlapi.ConnectDb(configParam.ConnectPostgres)
	if err != nil {
		log.Fatal(err)
		return
	}

	bot, err := tgbotapi.NewBotAPI(configParam.BotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		log.Printf(command)

		verification, err := verificationUser(update.Message.From)
		if err != nil {
			log.Println("sendMessage went wrong: ", err.Error())
		}

		if verification{
			verificationPigeon, err := checkingPigeonId(update.Message.From)

			if err != nil{
				log.Printf(err.Error())
			}

			if !verificationPigeon{
				err = ButtonAddNamePigeon(update.Message, bot)

				if err != nil {
					log.Printf(err.Error())
				}
			}

			checkingPigeonWork, err := sqlapi.CheckingPigeonWork(update.Message.From.UserName , sqlapi.DB)

			if err != nil{
				log.Printf(err.Error())
			}

			if !checkingPigeonWork{
				err = startPigeonWork(msg.BaseChat.ChatID, bot)			
			}

			if err != nil{
				log.Printf(err.Error())
			}

		} else {
			err = startRegisterUser(msg.BaseChat.ChatID, bot, command, update.Message)

			if err != nil {
				log.Printf(err.Error())
			}
		}




		if command == "" && !verification {
			err = buttonStart(msg.BaseChat.ChatID, bot)

		} else if command == ":bird: Start!" {

			err = ButtonAddedNewUser(update.Message, bot)

			if err != nil {
				log.Printf(err.Error())
			}

			err = ButtonAddNamePigeon(update.Message, bot)

			if err != nil {
				log.Printf(err.Error())
			}
			// err = emptyButton(msg.BaseChat.ChatID, bot)

			// if len(update.Message.Text) > 5 {
			// 	log.Printf(update.Message.Text)
			// }
			// //msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			// msg.ReplyToMessageID = update.Message.MessageID

			// err = sendMessage(msg.BaseChat.ChatID, "test message", bot, update.Message.MessageID)
			// if err != nil {
			// 	log.Println("sendMessage went wrong: ", err.Error())
			// }

			// bot.Send(msg)
		} else if command == "Задайте уникальное имя для вашего почтового голубя :bird:." {
			if len(msg.Text) >= 6 {
				refUP, err := saveRefUserPigeon(update.Message)
				if err != nil {
					log.Printf(err.Error())
				}
				if refUP == false {
					log.Printf("saveRefUserPigeon(update.Message)")
				}

			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Слишком короткое имя.")
				_, err := bot.Send(msg)
				if err != nil {
					log.Printf(err.Error())
				}
			}
		}

	}
}


func startRegisterUser(chatID int64, bot *tgbotapi.BotAPI, command string, message *tgbotapi.Message) error{

	if command == ""{
		err := buttonStart(chatID, bot)
		if err != nil{
			return err
		}

	} else if command == ":bird: Start!" {
		err := ButtonAddedNewUser(message, bot)
		if err != nil {
			log.Printf(err.Error())
			return err
		}

	} else{
		nameP := emoji.Sprint("Команда не распознана!")
		msg := tgbotapi.NewMessage(chatID, nameP)
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf(err.Error())
			return err
		}
	}
	return nil
}

func startPigeonWork(chatID int64, bot *tgbotapi.BotAPI) error{

	nameP := emoji.Sprint("Запустить работу сервиса?")

	msg := tgbotapi.NewMessage(chatID, nameP)

	startMessage := emoji.Sprint(":bird: Запустить!")
	stopMessage := emoji.Sprint(":bird: Остановить!")
	myKeyboardButtonStart := tgbotapi.KeyboardButton{Text: startMessage}
	myKeyboardButtonStop := tgbotapi.KeyboardButton{Text: stopMessage}
	// arrKeyboardButton := [][]tgbotapi.KeyboardButton{[]tgbotapi.KeyboardButton{myKeyboardButton}}
	arrKeyboardButton := [][]tgbotapi.KeyboardButton{{myKeyboardButtonStart, myKeyboardButtonStop}}
	replyKeyboardMarkup := tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        arrKeyboardButton,
		Selective:       true,
		OneTimeKeyboard: true,
	}

	msg.ReplyMarkup = replyKeyboardMarkup
	_, err := bot.Send(msg)
	command = ":bird: Start-Stop!"

	return err



}

func processingCommands(bot *tgbotapi.BotAPI, command string, message *tgbotapi.Message) error{

	if command == "Задайте уникальное имя для вашего почтового голубя :bird:." {
		if len(message.Text) >= 6 {
			refUP, err := saveRefUserPigeon(message)
			if err != nil {
				log.Printf(err.Error())
			}
			if refUP == false {
				log.Printf("saveRefUserPigeon(update.Message)")
			}

		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Слишком короткое имя.")
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf(err.Error())
			}
		}
	} 

	return nil
}


func sendMessage(chatID int64, message string, bot *tgbotapi.BotAPI, ReplyToMessageID int) error {

	msg := tgbotapi.NewMessage(chatID, message)

	// ttt := tgbotapi.NewKeyboardButton("Test button")

	// msg := tgbotapi.NewMessage(chatID, "", ttt)

	// _, err := bot.Send(ttt)

	myKeyboardButton := tgbotapi.KeyboardButton{Text: "Test keyboard"}
	arrKeyboardButton := []tgbotapi.KeyboardButton{myKeyboardButton}

	// newOneTimeReplyKeyboard := tgbotapi.NewOneTimeReplyKeyboard(arrKeyboardButton)

	newOneTimeReplyKeyboard := tgbotapi.NewReplyKeyboard(arrKeyboardButton)
	// markup := tgbotapi.InlineKeyboardMarkup{
	// 	InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
	// 		[]tgbotapi.InlineKeyboardButton{
	// 			tgbotapi.InlineKeyboardButton{Text: "Test"},
	// 		},
	// 	},
	// }

	// selectAction3key := tgbotapi.NewInlineKeyboardMarkup(
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData("1", "sa+1"),
	// 		tgbotapi.NewInlineKeyboardButtonData("2", "sa+2"),
	// 		tgbotapi.NewInlineKeyboardButtonData("3", "sa+3"),
	// 	),
	// )

	// markup2 := tgbotapi.ReplyKeyboardMarkup{
	// 	Keyboard: [][]tgbotapi.KeyboardButton{
	// 		[]tgbotapi.KeyboardButton{
	// 			tgbotapi.KeyboardButton{Text: "Test"},
	// 		},
	// 	},
	// }

	// edit := tgbotapi.NewEditMessageReplyMarkup(chatID, ReplyToMessageID, markup)

	//edit := tgbotapi.NewEditMessageReplyMarkup(chatID, ReplyToMessageID, selectAction3key)
	msg.ReplyMarkup = newOneTimeReplyKeyboard

	_, err := bot.Send(msg)

	// if edit.ReplyMarkup.InlineKeyboard[0][0].Text != "test" ||
	// 	edit.BaseEdit.ChatID != chatID ||
	// 	edit.BaseEdit.MessageID != ReplyToMessageID {
	// 	// t.Fail()
	// }

	return err
}

func buttonStart(chatID int64, bot *tgbotapi.BotAPI) error {
	nameP := emoji.Sprint("Get a personal pigeon?")

	msg := tgbotapi.NewMessage(chatID, nameP)

	startMessage := emoji.Sprint(":bird: Start!")
	myKeyboardButton := tgbotapi.KeyboardButton{Text: startMessage}
	// arrKeyboardButton := [][]tgbotapi.KeyboardButton{[]tgbotapi.KeyboardButton{myKeyboardButton}}
	arrKeyboardButton := [][]tgbotapi.KeyboardButton{{myKeyboardButton}}
	replyKeyboardMarkup := tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        arrKeyboardButton,
		Selective:       true,
		OneTimeKeyboard: true,
	}

	msg.ReplyMarkup = replyKeyboardMarkup
	_, err := bot.Send(msg)
	command = ":bird: Start!"

	return err
}

func buttonStop(chatID int64, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatID, "")

	myKeyboardButton := tgbotapi.KeyboardButton{Text: "Test keyboard"}
	arrKeyboardButton := []tgbotapi.KeyboardButton{myKeyboardButton}

	newOneTimeReplyKeyboard := tgbotapi.NewReplyKeyboard(arrKeyboardButton)
	msg.ReplyMarkup = newOneTimeReplyKeyboard

	_, err := bot.Send(msg)

	command = ":bird: Start!"

	return err
}

func emptyButton(chatID int64, bot *tgbotapi.BotAPI) error {

	msg := tgbotapi.NewMessage(chatID, emoji.Sprint("Come up with a name for the pigeon :bird:"))

	replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	msg.ReplyMarkup = replyKeyboardHide
	_, err := bot.Send(msg)

	return err
}

func verificationUser(user *tgbotapi.User) (bool, error) {

	userIn, err := sqlapi.GetUserByID(strconv.Itoa(user.ID), sqlapi.DB)

	if err != nil {
		log.Fatal(err)
		return false, err
	}

	if userIn == nil {
		return false, nil
	}

	return true, nil
}

func checkingPigeonId(user *tgbotapi.User) (bool, error){

	pigeonId, err := sqlapi.GetRefUserCodeByUserName(user.UserName, sqlapi.DB)
	if err != nil{
		log.Fatal(err)
			return false, err
	}

	if pigeonId == nil{
		return false, nil
	}

	return true, nil
}

// ButtonAddedNewUser ...
func ButtonAddedNewUser(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	saveUser, err := SaveNewUser(msg)

	if err != nil {
		return err
	}

	if saveUser == false {
		log.Fatal("Error save new user")
		return nil
	}
	return err


	// nameP := emoji.Sprint("Задайте уникальное имя для вашего почтового голубя :bird:. \nИмя должно состоять минимум из 6 символов")

	// m := tgbotapi.NewMessage(msg.Chat.ID, nameP)

	// replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	// m.ReplyMarkup = replyKeyboardHide
	// _, err = bot.Send(m)

	// command = "Задайте уникальное имя для вашего почтового голубя :bird:."
	// return err
}

func ButtonAddNamePigeon(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	nameP := emoji.Sprint("Задайте уникальное имя для вашего почтового голубя :bird:. \nИмя должно состоять минимум из 6 символов")

	m := tgbotapi.NewMessage(msg.Chat.ID, nameP)

	replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	m.ReplyMarkup = replyKeyboardHide
	_, err := bot.Send(m)

	command = "Задайте уникальное имя для вашего почтового голубя :bird:."
	return err
}

// SaveNewUser ...
func SaveNewUser(msg *tgbotapi.Message) (bool, error) {

	userInit, err := sqlapi.AddNewUser(msg.From.UserName,
		strconv.Itoa(msg.From.ID),
		strconv.Itoa(msg.From.ID),
		sqlapi.DB)
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	if userInit == nil {
		log.Fatal("Error save new user")
		return false, nil
	}

	return true, nil
}

func saveRefUserPigeon(msg *tgbotapi.Message) (bool, error) {
	pigeonName, err := sqlapi.AddNewCode(msg.Text, sqlapi.DB)

	if err != nil {
		log.Fatal(err)
		return false, err
	}
	if pigeonName == nil {
		return false, nil
	}

	addRefUserCode, err := sqlapi.AddRefUserCode(pigeonName.Code, strconv.Itoa(msg.From.ID), sqlapi.DB)

	if err != nil {
		log.Fatal(err)
		return false, err
	}
	if addRefUserCode == nil {
		return false, nil
	}

	return true, nil
}

// InitConfig ...
func InitConfig() (*config.Configuration, error) {
	var conf = &config.Configuration{}

	data, err := ioutil.ReadFile("./config/keys.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		fmt.Println("error: ", err)
		return nil, err
	}

	return conf, nil
}



Добавлять данные во временную БД
Проверять во временной БД состояние пользователя и переодически обновлять (если идет работа с ботом, 
							если нетработы бот дейтвует как устнаовлено)

Хранить во временной БД последнюю команду от пользователя (для отслеживания действий пользователя)
