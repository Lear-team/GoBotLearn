package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"
	"unicode/utf8"

	sqlapi "GoBotPigeon/app/sqlapi"
	config "GoBotPigeon/types"
	apitypes "GoBotPigeon/types/apitypes"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/kyokomi/emoji"
)

var configParam config.Configuration

var commandsBot = &config.Commands{}

func main() {

	configParam, err := initConfig()
	if err != nil {
		log.Panic(err)
	}

	err = sqlapi.ConnectDb(configParam.ConnectPostgres)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = getCommands()
	if err != nil {
		log.Panic(err)
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

		verification, err := verificationUser(update.Message.From)
		if err != nil {
			log.Println("sendMessage went wrong: ", err.Error())
		}

		lastCommand, err := sqlapi.GetLastCommandByUserName(update.Message.From.UserName, sqlapi.DB)
		if err != nil {
			log.Println(err.Error())
		}

		freshLastCommand, err := validLastCommand(lastCommand)

		if verification && freshLastCommand {
			if lastCommand.Command == "AddNamePigeon" {
				result, err := processingCommands(bot, update.Message)

				if err != nil {
					log.Printf(err.Error())
				}

				if result {
					err = buttonsStartPigeonWork(msg.BaseChat.ChatID, bot)

					if err != nil {
						log.Printf(err.Error())
					}
				} else {
					err = buttonAddNameBot(update.Message, bot)

					if err != nil {
						log.Printf(err.Error())
					}
				}
			} else if lastCommand.Command == "EditNamePigeon" {
				result, err := editUserCode(bot, update.Message)

				if err != nil {
					log.Printf(err.Error())
				}

				if result {
					err = buttonsStartPigeonWork(msg.BaseChat.ChatID, bot)

					if err != nil {
						log.Printf(err.Error())
					}
				} else {
					err = buttonAddNameBot(update.Message, bot)

					if err != nil {
						log.Printf(err.Error())
					}
				}
			}

		} else if verification {
			verificationPigeon, err := checkingPigeonID(update.Message.From)

			if err != nil {
				log.Printf(err.Error())
			}

			if !verificationPigeon {
				err = buttonAddNameBot(update.Message, bot)

				if err != nil {
					log.Printf(err.Error())
				}
			} else if verificationPigeon {
				checkingPigeonWork, err := sqlapi.CheckingPigeonWork(update.Message.From.UserName, sqlapi.DB)

				if err != nil {
					log.Printf(err.Error())
				}

				if !checkingPigeonWork {
					err = buttonsStartPigeonWork(msg.BaseChat.ChatID, bot)

					if err != nil {
						log.Printf(err.Error())
					}
				}

				if !checkingPigeonWork {

					if update.Message.Text == emoji.Sprint(commandsBot.StartBot) {
						err = sqlapi.StartPigeonWork(update.Message.From.UserName, sqlapi.DB)
						if err != nil {
							log.Printf(err.Error())
						}

						err = buttonsStopPigeonWork(msg.BaseChat.ChatID, bot)

						if err != nil {
							log.Printf(err.Error())
						}
					} else if update.Message.Text == emoji.Sprint(commandsBot.EditCode) {
						err = buttonEditNameBot(update.Message, bot)

						if err != nil {
							log.Printf(err.Error())
						}
					}

				} else if checkingPigeonWork && update.Message.Text == emoji.Sprint(commandsBot.StopBot) {

					err = sqlapi.StopPigeonWork(update.Message.From.UserName, sqlapi.DB)
					if err != nil {
						log.Printf(err.Error())
					}

					err = buttonsStartPigeonWork(msg.BaseChat.ChatID, bot)
					if err != nil {
						log.Printf(err.Error())
					}

				} else if checkingPigeonWork {
					err = buttonsStopPigeonWork(msg.BaseChat.ChatID, bot)

					if err != nil {
						log.Printf(err.Error())
					}
				}
			}

			if err != nil {
				log.Printf(err.Error())
			}

		} else {
			command := emoji.Sprint(commandsBot.StartBot)
			if update.Message.Text == command {
				err = startRegisterUser(msg.BaseChat.ChatID, bot, update.Message)

				if err != nil {
					log.Printf(err.Error())
				}
			} else {
				err = buttonStart(msg.BaseChat.ChatID, bot)
				if err != nil {
					log.Printf(err.Error())
				}
			}
		}
	}
}

func startRegisterUser(chatID int64, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {

	err := buttonAddedNewUser(message, bot)
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	// продолжение работы с юзером, предложить создать кодовое слово
	err = buttonAddNameBot(message, bot)

	if err != nil {
		log.Printf(err.Error())
	}
	return nil
}

func buttonsStartPigeonWork(chatID int64, bot *tgbotapi.BotAPI) error {

	nameP := emoji.Sprint("Запустить работу сервиса?")

	msg := tgbotapi.NewMessage(chatID, nameP)

	startMessage := emoji.Sprint(":bird: Запустить!")
	changeMessage := emoji.Sprint(":bird: Сменить имя боту.")
	myKeyboardButtonStart := tgbotapi.KeyboardButton{Text: startMessage}
	myKeyboardButtonStop := tgbotapi.KeyboardButton{Text: changeMessage}
	arrKeyboardButton := [][]tgbotapi.KeyboardButton{{myKeyboardButtonStart, myKeyboardButtonStop}}
	replyKeyboardMarkup := tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        arrKeyboardButton,
		Selective:       true,
		OneTimeKeyboard: true,
	}

	msg.ReplyMarkup = replyKeyboardMarkup
	_, err := bot.Send(msg)

	return err
}

func buttonsStopPigeonWork(chatID int64, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatID, "Бот запущен.")

	stopMessage := emoji.Sprint(":bird: Остановить!")
	myKeyboardButtonStop := tgbotapi.KeyboardButton{Text: stopMessage}
	arrKeyboardButton := [][]tgbotapi.KeyboardButton{{myKeyboardButtonStop}}
	replyKeyboardMarkup := tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        arrKeyboardButton,
		Selective:       true,
		OneTimeKeyboard: true,
	}

	msg.ReplyMarkup = replyKeyboardMarkup
	_, err := bot.Send(msg)
	return err
}

func processingCommands(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (bool, error) {

	var rune = utf8.RuneCountInString(message.Text)
	if len(message.Text) >= 6 && rune >= 6 {
		refUP, err := saveRefUserPigeon(message)
		if err != nil {
			log.Printf(err.Error())
			return false, err
		}
		if refUP == false {
			log.Printf("saveRefUserPigeon(update.Message)")
		}

		if refUP == true {
			err = sqlapi.DeleteLastCommand(message.From.UserName, "AddNamePigeon", sqlapi.DB)
		}

		if err != nil {
			log.Printf(err.Error())
			return false, err
		}

		return refUP, nil
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Слишком короткое имя.")
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf(err.Error())
		return false, err
	}

	return false, nil
}

func editUserCode(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (bool, error) {
	var rune = utf8.RuneCountInString(message.Text)
	if len(message.Text) >= 6 && rune >= 6 {
		refUP, err := updateRefUserCode(message)
		if err != nil {
			log.Printf(err.Error())
			return false, err
		}
		if refUP == false {
			log.Printf("updateRefUserCode(update.Message)")
		}

		if refUP == true {
			err = sqlapi.DeleteLastCommand(message.From.UserName, "EditNamePigeon", sqlapi.DB)
		}

		if err != nil {
			log.Printf(err.Error())
			return false, err
		}

		return refUP, nil
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Слишком короткое имя.")
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf(err.Error())
		return false, err
	}

	return false, nil
}

func sendMessage(chatID int64, message string, bot *tgbotapi.BotAPI, ReplyToMessageID int) error {

	msg := tgbotapi.NewMessage(chatID, message)
	myKeyboardButton := tgbotapi.KeyboardButton{Text: "Test keyboard"}
	arrKeyboardButton := []tgbotapi.KeyboardButton{myKeyboardButton}
	newOneTimeReplyKeyboard := tgbotapi.NewReplyKeyboard(arrKeyboardButton)
	msg.ReplyMarkup = newOneTimeReplyKeyboard
	_, err := bot.Send(msg)
	return err
}

func buttonStart(chatID int64, bot *tgbotapi.BotAPI) error {
	nameP := emoji.Sprint("Создать персонального бота?")

	msg := tgbotapi.NewMessage(chatID, nameP)

	startMessage := emoji.Sprint(commandsBot.StartBot)
	myKeyboardButton := tgbotapi.KeyboardButton{Text: startMessage}
	arrKeyboardButton := [][]tgbotapi.KeyboardButton{{myKeyboardButton}}
	replyKeyboardMarkup := tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        arrKeyboardButton,
		Selective:       true,
		OneTimeKeyboard: true,
	}

	msg.ReplyMarkup = replyKeyboardMarkup
	_, err := bot.Send(msg)

	return err
}

func buttonStop(chatID int64, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatID, "")

	myKeyboardButton := tgbotapi.KeyboardButton{Text: "Test keyboard"}
	arrKeyboardButton := []tgbotapi.KeyboardButton{myKeyboardButton}

	newOneTimeReplyKeyboard := tgbotapi.NewReplyKeyboard(arrKeyboardButton)
	msg.ReplyMarkup = newOneTimeReplyKeyboard

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

func checkingPigeonID(user *tgbotapi.User) (bool, error) {

	pigeonID, err := sqlapi.GetRefUserCodeByUserName(user.UserName, sqlapi.DB)
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	if pigeonID == nil {
		return false, nil
	}

	return true, nil
}

func buttonAddedNewUser(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	saveUser, err := saveNewUser(msg)

	if err != nil {
		return err
	}

	if saveUser == false {
		log.Fatal("Error save new user")
		return nil
	}
	return err
}

func buttonAddNameBot(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	nameP := emoji.Sprint("Задайте уникальное имя для вашего почтового бота :bird:. \nИмя должно состоять минимум из 6 символов")

	m := tgbotapi.NewMessage(msg.Chat.ID, nameP)

	replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	m.ReplyMarkup = replyKeyboardHide
	_, err := bot.Send(m)

	err = sqlapi.SetLastComandUser(msg.From.UserName, sqlapi.DB, "AddNamePigeon")

	return err
}

func buttonEditNameBot(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	nameP := emoji.Sprint("Задайте уникальное имя для вашего почтового бота :bird:. \nИмя должно состоять минимум из 6 символов")

	m := tgbotapi.NewMessage(msg.Chat.ID, nameP)

	replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	m.ReplyMarkup = replyKeyboardHide
	_, err := bot.Send(m)

	err = sqlapi.SetLastComandUser(msg.From.UserName, sqlapi.DB, "EditNamePigeon")

	return err
}

func saveNewUser(msg *tgbotapi.Message) (bool, error) {

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

func updateRefUserCode(msg *tgbotapi.Message) (bool, error) {
	update, err := sqlapi.UpdateRefUserCode(msg.Text, msg.From.UserName, sqlapi.DB)

	if err != nil {
		log.Fatal(err)
		return false, err
	}
	if update == nil {
		return false, nil
	}

	return true, nil
}

func validLastCommand(lastCommand *apitypes.LastUserCommand) (bool, error) {

	today := time.Now()
	t := today.Format("2006/1/2 15:04")

	if lastCommand != nil && lastCommand.Command == "AddNamePigeon" {
		if lastCommand.DataCommand.Add(10*time.Minute).Format("2006/1/2 15:04") > t {
			return true, nil
		}
	} else if lastCommand != nil && lastCommand.Command == "EditNamePigeon" {
		if lastCommand.DataCommand.Add(10*time.Minute).Format("2006/1/2 15:04") > t {
			return true, nil
		}
	}

	return false, nil
}

func initConfig() (*config.Configuration, error) {
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

func getCommands() error {

	data, err := ioutil.ReadFile("./config/command.json")

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &commandsBot)
	if err != nil {
		fmt.Println("error: ", err)
		return err
	}

	return nil
}
