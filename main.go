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
		if command == "" && !verification {
			err = buttonStart(msg.BaseChat.ChatID, bot)
		} else if command == ":bird: Start!" {
			err = emptyButton(msg.BaseChat.ChatID, bot)

			if len(update.Message.Text) > 5 {
				log.Printf(update.Message.Text)
			}
			//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			err = sendMessage(msg.BaseChat.ChatID, "test message", bot, update.Message.MessageID)
			if err != nil {
				log.Println("sendMessage went wrong: ", err.Error())
			}

			bot.Send(msg)
		}

	}
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
	// msg := tgbotapi.NewMessage(chatID, "eee")

	// selectAction3key := tgbotapi.NewInlineKeyboardMarkup(
	// 	tgbotapi.NewInlineKeyboardRow(
	// 		tgbotapi.NewInlineKeyboardButtonData("1", "sa+1"),
	// 		tgbotapi.NewInlineKeyboardButtonData("2", "sa+2"),
	// 		tgbotapi.NewInlineKeyboardButtonData("3", "sa+3"),
	// 	),
	// )
	startMessage := emoji.Sprint(":bird: Start!")
	myKeyboardButton := tgbotapi.KeyboardButton{Text: startMessage}
	// myKeyboardButton2 := tgbotapi.KeyboardButton{Text: "Stop"}

	// arrKeyboardButton := [][]tgbotapi.KeyboardButton{[]tgbotapi.KeyboardButton{myKeyboardButton, myKeyboardButton2}}
	arrKeyboardButton := [][]tgbotapi.KeyboardButton{[]tgbotapi.KeyboardButton{myKeyboardButton}}

	// newOneTimeReplyKeyboard := tgbotapi.NewReplyKeyboard(arrKeyboardButton)
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