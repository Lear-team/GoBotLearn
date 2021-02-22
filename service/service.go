package service

import (
	"log"
	"strconv"
	"time"
	"unicode/utf8"

	"GoBotPigeon/app/sqlapi"
	"GoBotPigeon/types"
	"GoBotPigeon/types/apitypes"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/pkg/errors"
)

type Storage interface {
	GetLastCommandByUserName(usename string) (*apitypes.LastUserCommand, error)
	CheckingPigeonWork(userN string) (bool, error)
	StopPigeonWork(userN string) error
	AddNewUser2(username string) error
}

type BotSvc struct {
	storage     Storage
	commandsBot types.Commands
}

func NewBotSvc(s Storage, commandsBot types.Commands) *BotSvc {
	return &BotSvc{
		storage:     s,
		commandsBot: commandsBot,
	}
}

func (b *BotSvc) NewMethod(newUserName string) error {
	if err := b.storage.AddNewUser2(newUserName); err != nil {
		return nil
	}
	return nil
}

func (b *BotSvc) ProcessingCommands(message *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)

	verification, err := b.verificationUser(message.From)
	if err != nil {
		log.Println("User verification failed with an error: ", err.Error())
		return err
	}

	lastCommand, err := b.storage.GetLastCommandByUserName(message.From.UserName)
	if err != nil {
		return errors.Wrap(err, "Getting the last command failed")
	}

	freshLastCommand, err := b.validLastCommand(lastCommand)
	if err != nil {
		log.Println("Checking the last command failed with an error: ", err.Error())
		return err
	}

	if verification && freshLastCommand {
		err = b.verificationFreshLastCommand(message, bot, msg, lastCommand.Command)
		if err != nil {
			log.Println("verificationFreshLastCommand failed with an error: ", err.Error())
			return err
		}
	} else if verification {

		err := b.verificationWorkBot(message, bot, msg)
		if err != nil {
			log.Println("verificationWorkBot failed with an error: ", err.Error())
			return err
		}

	} else {
		err = b.offerRregister(message, bot, msg)
		if err != nil {
			log.Printf(err.Error())
			return err
		}
	}

	return nil
}

func (b *BotSvc) verificationFreshLastCommand(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig, command string) error {
	if command == b.commandsBot.AddNameBot {
		err := b.commandAddNameBot(message, bot, msg)
		if err != nil {
			log.Printf(err.Error())
			return err
		}

	} else if command == b.commandsBot.EditNameBot {
		err := b.commandEditNameBot(message, bot, msg)
		if err != nil {
			log.Printf(err.Error())
			return err
		}
	}
	return nil
}

func (b *BotSvc) commandAddNameBot(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	result, err := b.addUserCode(bot, message)

	if err != nil {
		log.Println("Bot name creation failed with an error: ", err.Error())
		return err
	}

	if result {
		err = b.buttonsStartBotWork(msg.BaseChat.ChatID, bot)

		if err != nil {
			log.Println("Bot name creation failed with an error: ", err.Error())
			return err
		}
	} else {
		err = b.buttonAddNameBot(message, bot)

		if err != nil {
			log.Println("Adding name bot failed with an error: ", err.Error())
			return err
		}
	}
	return nil
}

func (b *BotSvc) commandEditNameBot(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	result, err := b.editUserCode(bot, message)

	if err != nil {
		log.Println("Updating user code failed with an error: ", err.Error())
		return err
	}

	if result {
		err = b.buttonsStartBotWork(msg.BaseChat.ChatID, bot)

		if err != nil {
			log.Println("Starting work failed with an error: ", err.Error())
			return err
		}
	} else {
		err = b.buttonAddNameBot(message, bot)

		if err != nil {
			log.Println("Adding name bot failed with an error: ", err.Error())
			return err
		}
	}
	return nil
}

func (b *BotSvc) verificationWorkBot(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	checkingBotWork, err := b.storage.CheckingPigeonWork(message.From.UserName)

	if err != nil {
		log.Println("CheckingPigeonWork failed with an error: ", err.Error())
		return err
	}

	if !checkingBotWork {

		err = b.botDoNotWork(message, bot, msg)
		if err != nil {
			log.Println("botDoNotWork failed with an error: ", err.Error())
			return err
		}

	} else if checkingBotWork && message.Text == emoji.Sprint(b.commandsBot.StopBot) {

		err = b.stopBotWorking(message, bot, msg)
		if err != nil {
			log.Println("botDoNotWork failed with an error: ", err.Error())
			return err
		}

	} else if checkingBotWork {
		err = b.botWork(bot, msg)
		if err != nil {
			log.Println("botWork failed with an error: ", err.Error())
			return err
		}
	}

	return err
}

func (b *BotSvc) botDoNotWork(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	if message.Text == emoji.Sprint(b.commandsBot.StartBot) {
		err := sqlapi.StartPigeonWork(message.From.UserName, sqlapi.DB)
		if err != nil {
			log.Println("StartPigeonWork failed with an error: ", err.Error())
			return err
		}

		err = b.buttonsStopBotWork(msg.BaseChat.ChatID, bot)

		if err != nil {
			log.Println("buttonsStopBotWork failed with an error: ", err.Error())
			return err
		}
	} else if message.Text == emoji.Sprint(b.commandsBot.EditCode) {
		err := b.buttonEditNameBot(message, bot)

		if err != nil {
			log.Println("buttonEditNameBot failed with an error: ", err.Error())
			return err
		}
	}
	return nil
}

func (b *BotSvc) botWork(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	err := b.buttonsStopBotWork(msg.BaseChat.ChatID, bot)

	if err != nil {
		log.Println("buttonsStopBotWork failed with an error: ", err.Error())
		return err
	}
	return nil
}

func (b *BotSvc) stopBotWorking(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	err := b.storage.StopPigeonWork(message.From.UserName)
	if err != nil {
		log.Println("StopPigeonWork failed with an error: ", err.Error())
		return err
	}

	err = b.buttonsStartBotWork(msg.BaseChat.ChatID, bot)
	if err != nil {
		log.Println("buttonsStartBotWork failed with an error: ", err.Error())
		return err
	}

	return nil
}

func (b *BotSvc) offerRregister(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	command := emoji.Sprint(b.commandsBot.StartBot)
	if message.Text == command {
		err := b.startRegisterUser(msg.BaseChat.ChatID, bot, message)

		if err != nil {
			log.Printf(err.Error())
			return err
		}
	} else {
		err := b.buttonStart(msg.BaseChat.ChatID, bot)
		if err != nil {
			log.Printf(err.Error())
			return err
		}
	}
	return nil
}

func (b *BotSvc) startRegisterUser(chatID int64, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	if err := b.buttonAddedNewUser(message, bot); err != nil {
		log.Printf(err.Error())
		return err
	}

	// продолжение работы с юзером, предложить создать кодовое слово
	if err := b.buttonAddNameBot(message, bot); err != nil {
		log.Printf(err.Error())
	}

	return nil
}

func (b *BotSvc) buttonsStartBotWork(chatID int64, bot *tgbotapi.BotAPI) error {
	arrKeyboardButton := [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.KeyboardButton{
				Text: emoji.Sprint(b.commandsBot.StartBot),
			},
			tgbotapi.KeyboardButton{
				Text: emoji.Sprint(b.commandsBot.EditCode),
			},
		},
	}
	replyKeyboardMarkup := tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        arrKeyboardButton,
		Selective:       true,
		OneTimeKeyboard: true,
	}

	msg := tgbotapi.NewMessage(chatID, emoji.Sprint("Запустить работу сервиса?"))
	msg.ReplyMarkup = replyKeyboardMarkup

	if _, err := bot.Send(msg); err != nil {
		return err
	}

	return nil
}

func (b *BotSvc) buttonsStopBotWork(chatID int64, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatID, "Бот запущен.")

	stopMessage := emoji.Sprint(b.commandsBot.StopBot)
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

func (b *BotSvc) addUserCode(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (bool, error) {

	var rune = utf8.RuneCountInString(message.Text)
	if len(message.Text) >= 6 && rune >= 6 {
		refUP, err := b.saveRefUserCode(message)
		if err != nil {
			log.Println("Saving the bot name failed with an error: ", err.Error())
			return false, err
		}

		if refUP == false {
			log.Printf("Bot name not saved.")
		}

		if refUP == true {
			err = sqlapi.DeleteLastCommand(message.From.UserName, b.commandsBot.AddNameBot, sqlapi.DB)
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

func (b *BotSvc) editUserCode(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (bool, error) {

	var rune = utf8.RuneCountInString(message.Text)
	if len(message.Text) >= 6 && rune >= 6 {
		refUP, err := b.updateRefUserCode(message)
		if err != nil {
			log.Println("Editing  the bot name failed with an error: ", err.Error())
			return false, err
		}
		if refUP == false {
			log.Printf("Bot name not updated.")
		}

		if refUP == true {
			err = sqlapi.DeleteLastCommand(message.From.UserName, b.commandsBot.EditNameBot, sqlapi.DB)
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

func (b *BotSvc) buttonStart(chatID int64, bot *tgbotapi.BotAPI) error {
	nameP := emoji.Sprint("Создать персонального бота?")

	msg := tgbotapi.NewMessage(chatID, nameP)

	startMessage := emoji.Sprint(b.commandsBot.StartBot)
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

func (b *BotSvc) verificationUser(user *tgbotapi.User) (bool, error) {

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

func (b *BotSvc) checkingPigeonID(user *tgbotapi.User) (bool, error) {

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

func (b *BotSvc) buttonAddedNewUser(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	saveUser, err := b.saveNewUser(msg)

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

func (b *BotSvc) buttonAddNameBot(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	nameP := emoji.Sprint("Задайте уникальное имя для вашего почтового бота :bird:. \nИмя должно состоять минимум из 6 символов")

	m := tgbotapi.NewMessage(msg.Chat.ID, nameP)

	replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	m.ReplyMarkup = replyKeyboardHide
	_, err := bot.Send(m)
	if err != nil {
		log.Println("Sending the message failed with an error: ", err.Error())
		return err
	}

	err = sqlapi.SetLastComandUser(msg.From.UserName, sqlapi.DB, b.commandsBot.AddNameBot)
	if err != nil {
		log.Println("SetLastComandUser failed with an error: ", err.Error())
		return err
	}

	return err
}

func (b *BotSvc) buttonEditNameBot(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	nameP := emoji.Sprint("Задайте уникальное имя для вашего почтового бота :bird:. \nИмя должно состоять минимум из 6 символов")

	m := tgbotapi.NewMessage(msg.Chat.ID, nameP)

	replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	m.ReplyMarkup = replyKeyboardHide
	_, err := bot.Send(m)
	if err != nil {
		log.Println("Sending the message failed with an error: ", err.Error())
		return err
	}

	err = sqlapi.SetLastComandUser(msg.From.UserName, sqlapi.DB, b.commandsBot.EditNameBot)
	if err != nil {
		log.Println("SetLastComandUser failed with an error: ", err.Error())
		return err
	}

	return err
}

func (b *BotSvc) saveNewUser(msg *tgbotapi.Message) (bool, error) {

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

func (b *BotSvc) saveRefUserCode(msg *tgbotapi.Message) (bool, error) {
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

func (b *BotSvc) updateRefUserCode(msg *tgbotapi.Message) (bool, error) {
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

func (b *BotSvc) validLastCommand(lastCommand *apitypes.LastUserCommand) (bool, error) {

	today := time.Now()
	t := today.Format("2006/1/2 15:04")

	if lastCommand != nil && lastCommand.Command == b.commandsBot.AddNameBot {
		if lastCommand.DataCommand.Add(10*time.Minute).Format("2006/1/2 15:04") > t {
			return true, nil
		}
	} else if lastCommand != nil && lastCommand.Command == b.commandsBot.EditNameBot {
		if lastCommand.DataCommand.Add(10*time.Minute).Format("2006/1/2 15:04") > t {
			return true, nil
		}
	}

	return false, nil
}
