package main

import (
	"encoding/json"
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

var configParam = &config.Configuration{}

var commandsBot = &config.Commands{}

func main() {

	err := initConfig()
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
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(configParam.BotToken)
	if err != nil {
		log.Fatal(err)
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

		err = processingCommands(update.Message, bot)
		if err != nil {
			log.Println("Processing commands: ", err.Error())
		}
	}
}

func processingCommands(message *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)

	verification, err := verificationUser(message.From)
	if err != nil {
		log.Println("User verification failed with an error: ", err.Error())
		return err
	}

	lastCommand, err := sqlapi.GetLastCommandByUserName(message.From.UserName, sqlapi.DB)
	if err != nil {
		log.Println("Getting the last command failed with an error: ", err.Error())
		return err
	}

	freshLastCommand, err := validLastCommand(lastCommand)
	if err != nil {
		log.Println("Checking the last command failed with an error: ", err.Error())
		return err
	}

	if verification && freshLastCommand {
		err = verificationFreshLastCommand(message, bot, msg, lastCommand.Command)
		if err != nil {
			log.Printf(err.Error())
			return err
		}
	} else if verification {
		verificationPigeon, err := checkingPigeonID(message.From)

		if err != nil {
			log.Println("Checking Pigeon ID failed with an error: ", err.Error())
			return err
		}

		if !verificationPigeon {
			err = buttonAddNameBot(message, bot)

			if err != nil {
				log.Println("Sending the message failed with an error: ", err.Error())
				return err
			}
		} else if verificationPigeon {
			checkingPigeonWork, err := sqlapi.CheckingPigeonWork(message.From.UserName, sqlapi.DB)

			if err != nil {
				log.Printf(err.Error())
			}

			if !checkingPigeonWork {

				if message.Text == emoji.Sprint(commandsBot.StartBot) {
					err = sqlapi.StartPigeonWork(message.From.UserName, sqlapi.DB)
					if err != nil {
						log.Printf(err.Error())
					}

					err = buttonsStopPigeonWork(msg.BaseChat.ChatID, bot)

					if err != nil {
						log.Printf(err.Error())
					}
				} else if message.Text == emoji.Sprint(commandsBot.EditCode) {
					err = buttonEditNameBot(message, bot)

					if err != nil {
						log.Printf(err.Error())
					}
				}

			} else if checkingPigeonWork && message.Text == emoji.Sprint(commandsBot.StopBot) {

				err = sqlapi.StopPigeonWork(message.From.UserName, sqlapi.DB)
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
		err = offerRregister(message, bot, msg)
		if err != nil {
			log.Printf(err.Error())
			return err
		}
	}

	return nil
}

func verificationFreshLastCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig, command string) error {
	if command == commandsBot.AddNameBot {
		result, err := addUserCode(bot, message)

		if err != nil {
			log.Println("Bot name creation failed with an error: ", err.Error())
			return err
		}

		if result {
			err = buttonsStartPigeonWork(msg.BaseChat.ChatID, bot)

			if err != nil {
				log.Println("Bot name creation failed with an error: ", err.Error())
				return err
			}
		} else {
			err = buttonAddNameBot(message, bot)

			if err != nil {
				log.Println("Adding name bot failed with an error: ", err.Error())
				return err
			}
		}
	} else if command == commandsBot.EditNameBot {
		result, err := editUserCode(bot, message)

		if err != nil {
			log.Println("Updating user code failed with an error: ", err.Error())
			return err
		}

		if result {
			err = buttonsStartPigeonWork(msg.BaseChat.ChatID, bot)

			if err != nil {
				log.Println("Starting work failed with an error: ", err.Error())
				return err
			}
		} else {
			err = buttonAddNameBot(message, bot)

			if err != nil {
				log.Println("Adding name bot failed with an error: ", err.Error())
				return err
			}
		}
	}
	return nil
}

func offerRregister(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	command := emoji.Sprint(commandsBot.StartBot)
	if message.Text == command {
		err := startRegisterUser(msg.BaseChat.ChatID, bot, message)

		if err != nil {
			log.Printf(err.Error())
			return err
		}
	} else {
		err := buttonStart(msg.BaseChat.ChatID, bot)
		if err != nil {
			log.Printf(err.Error())
			return err
		}
	}
	return nil
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

	startMessage := emoji.Sprint(commandsBot.StartBot)
	changeMessage := emoji.Sprint(commandsBot.EditCode)
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

	stopMessage := emoji.Sprint(commandsBot.StopBot)
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

func addUserCode(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (bool, error) {

	var rune = utf8.RuneCountInString(message.Text)
	if len(message.Text) >= 6 && rune >= 6 {
		refUP, err := saveRefUserCode(message)
		if err != nil {
			log.Println("Saving the bot name failed with an error: ", err.Error())
			return false, err
		}

		if refUP == false {
			log.Printf("Bot name not saved.")
		}

		if refUP == true {
			err = sqlapi.DeleteLastCommand(message.From.UserName, commandsBot.AddNameBot, sqlapi.DB)
			if err != nil {
				log.Println("Delete last command failed with an error: ", err.Error())
				return false, err
			}
		}
		return refUP, nil
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Слишком короткое имя.")
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Sending the message failed with an error: ", err.Error())
		return false, err
	}

	return false, nil
}

func editUserCode(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (bool, error) {

	var rune = utf8.RuneCountInString(message.Text)
	if len(message.Text) >= 6 && rune >= 6 {
		refUP, err := updateRefUserCode(message)
		if err != nil {
			log.Println("Editing  the bot name failed with an error: ", err.Error())
			return false, err
		}
		if refUP == false {
			log.Printf("Bot name not updated.")
		}

		if refUP == true {
			err = sqlapi.DeleteLastCommand(message.From.UserName, commandsBot.EditNameBot, sqlapi.DB)
			if err != nil {
				log.Println("Delete last command failed with an error: ", err.Error())
				return false, err
			}
		}
		return refUP, nil
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Слишком короткое имя.")
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Sending the message failed with an error: ", err.Error())
		return false, err
	}

	return false, nil
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
	if err != nil {
		log.Println("Sending the message failed with an error: ", err.Error())
		return err
	}

	return err
}

func verificationUser(user *tgbotapi.User) (bool, error) {

	userIn, err := sqlapi.GetUserByID(strconv.Itoa(user.ID), sqlapi.DB)

	if err != nil {
		log.Println("GetUserByID failed with an error: ", err.Error())
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
		log.Println("GetRefUserCodeByUserName failed with an error: ", err.Error())
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
		log.Println("Saving new user failed with an error: ", err.Error())
		return err
	}

	if saveUser == false {
		log.Println("The new user is not saved")
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
	if err != nil {
		log.Println("Sending the message failed with an error: ", err.Error())
		return err
	}

	err = sqlapi.SetLastComandUser(msg.From.UserName, sqlapi.DB, commandsBot.AddNameBot)
	if err != nil {
		log.Println("SetLastComandUser failed with an error: ", err.Error())
		return err
	}

	return err
}

func buttonEditNameBot(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	nameP := emoji.Sprint("Задайте уникальное имя для вашего почтового бота :bird:. \nИмя должно состоять минимум из 6 символов")

	m := tgbotapi.NewMessage(msg.Chat.ID, nameP)

	replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	m.ReplyMarkup = replyKeyboardHide
	_, err := bot.Send(m)
	if err != nil {
		log.Println("Sending the message failed with an error: ", err.Error())
		return err
	}

	err = sqlapi.SetLastComandUser(msg.From.UserName, sqlapi.DB, commandsBot.EditNameBot)
	if err != nil {
		log.Println("SetLastComandUser failed with an error: ", err.Error())
		return err
	}

	return err
}

func saveNewUser(msg *tgbotapi.Message) (bool, error) {

	userInit, err := sqlapi.AddNewUser(msg.From.UserName,
		strconv.Itoa(msg.From.ID),
		strconv.Itoa(msg.From.ID),
		sqlapi.DB)
	if err != nil {
		log.Println("Adding new user failed with an error: ", err.Error())
		return false, err
	}

	if userInit == nil {
		return false, nil
	}

	return true, nil
}

func saveRefUserCode(msg *tgbotapi.Message) (bool, error) {
	botName, err := sqlapi.AddNewCode(msg.Text, sqlapi.DB)

	if err != nil {
		log.Println("Adding new code failed with an error: ", err.Error())
		return false, err
	}
	if botName == nil {
		return false, nil
	}

	addRefUserCode, err := sqlapi.AddRefUserCode(botName.Code, strconv.Itoa(msg.From.ID), sqlapi.DB)

	if err != nil {
		log.Println("Adding RefUserCode failed with an error: ", err.Error())
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
		log.Println("Updating RefUserCode failed with an error: ", err.Error())
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

	if lastCommand != nil && lastCommand.Command == commandsBot.AddNameBot {
		if lastCommand.DataCommand.Add(10*time.Minute).Format("2006/1/2 15:04") > t {
			return true, nil
		}
	} else if lastCommand != nil && lastCommand.Command == commandsBot.EditNameBot {
		if lastCommand.DataCommand.Add(10*time.Minute).Format("2006/1/2 15:04") > t {
			return true, nil
		}
	}

	return false, nil
}

func initConfig() error {

	data, err := ioutil.ReadFile("./config/keys.json")
	if err != nil {
		log.Println("Reading the configuration file ended with an error: ", err.Error())
		return err
	}

	err = json.Unmarshal(data, &configParam)
	if err != nil {
		log.Println("Unmarshal ended with an error: ", err.Error())
		return err
	}

	return nil
}

func getCommands() error {

	data, err := ioutil.ReadFile("./config/command.json")
	if err != nil {
		log.Println("Reading the command file ended with an error: ", err.Error())
		return err
	}

	err = json.Unmarshal(data, &commandsBot)
	if err != nil {
		log.Println("Unmarshal ended with an error: ", err.Error())
		return err
	}

	return nil
}
