package service

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/pkg/errors"
)

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
		return errors.Wrap(err, "bot.Send failed")
	}

	return nil
}

func (b *BotSvc) buttonAddNameBot(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	nameP := emoji.Sprint("Задайте уникальное имя для вашего почтового бота :bird:. \nИмя должно состоять минимум из 6 символов")

	m := tgbotapi.NewMessage(msg.Chat.ID, nameP)

	replyKeyboardHide := tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

	m.ReplyMarkup = replyKeyboardHide
	_, err := bot.Send(m)
	if err != nil {
		return errors.Wrap(err, "Sending the message failed")
	}

	err = b.storage.SetLastComandUser(context.Background(), msg.From.UserName, b.commandsBot.AddNameBot)
	if err != nil {
		return errors.Wrap(err, "SetLastComandUser failed")
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
		return errors.Wrap(err, "Sending the message failed")
	}

	err = b.storage.SetLastComandUser(context.Background(), msg.From.UserName, b.commandsBot.EditNameBot)
	if err != nil {
		return errors.Wrap(err, "SetLastComandUser failed")
	}

	return err
}

func (b *BotSvc) buttonAddedNewUser(msg *tgbotapi.Message, bot *tgbotapi.BotAPI) error {

	saveUser, err := b.saveNewUser(msg)

	if err != nil {
		return errors.Wrap(err, "Saving new user failed")
	}

	if saveUser == false {
		log.Println("The new user is not saved")
		return nil
	}
	return err
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
		return errors.Wrap(err, "Sending the message failed")
	}

	return err
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
	if _, err := bot.Send(msg); err != nil {
		return errors.Wrap(err, "bot.Send failed")
	}
	return nil
}
