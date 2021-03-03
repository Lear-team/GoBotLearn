package service

import (
	"context"
	"strconv"
	"time"

	"GoBotPigeon/types"
	"GoBotPigeon/types/apitypes"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/pkg/errors"
)

func (b *BotSvc) GetVerificationFlags(message *tgbotapi.Message) (types.VerificationFlags, error) {

	var verificationFlags types.VerificationFlags

	verification, err := b.verificationUser(message.From)
	if err != nil {
		return verificationFlags, errors.Wrap(err, "User verification failed")
	}

	lastCommand, err := b.storage.GetLastCommandByUserName(context.Background(), message.From.UserName)
	if err != nil {
		return verificationFlags, errors.Wrap(err, "Getting the last command failed")
	}

	freshLastCommand, err := b.validLastCommand(lastCommand)
	if err != nil {
		return verificationFlags, errors.Wrap(err, "Checking the last command failed")
	}

	verificationFlags.FreshLastCommand = freshLastCommand
	verificationFlags.Verification = verification
	verificationFlags.LastCommand = lastCommand

	return verificationFlags, nil
}

func (b *BotSvc) verificationUser(user *tgbotapi.User) (bool, error) {

	userIn, err := b.storage.GetUserByID(context.Background(), strconv.Itoa(user.ID))

	if err != nil {
		return false, errors.Wrap(err, "GetUserByID failed")
	}

	return userIn != nil, nil
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

func (b *BotSvc) verificationWorkBot(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	checkingBotWork, err := b.storage.CheckingPigeonWork(context.Background(), strconv.Itoa(message.From.ID))

	if err != nil {
		return errors.Wrap(err, "CheckingPigeonWork failed")
	}

	if !checkingBotWork {

		err = b.botDoNotWork(message, bot, msg)
		if err != nil {
			return errors.Wrap(err, "botDoNotWork failed")
		}

	} else if checkingBotWork && message.Text == emoji.Sprint(b.commandsBot.StopBot) {

		err = b.stopBotWorking(message, bot, msg)
		if err != nil {
			return errors.Wrap(err, "stopBotWorking failed")
		}

	} else if checkingBotWork {
		err = b.botWork(bot, msg)
		if err != nil {
			return errors.Wrap(err, "botWork failed")
		}
	}

	return err
}

func (b *BotSvc) botWork(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	err := b.buttonsStopBotWork(msg.BaseChat.ChatID, bot)

	if err != nil {
		return errors.Wrap(err, "buttonsStopBotWork failed")
	}
	return nil
}
