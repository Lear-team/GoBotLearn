package service

import (
	"context"
	"log"
	"strconv"
	"unicode/utf8"

	"GoBotPigeon/types"
	"GoBotPigeon/types/apitypes"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kyokomi/emoji"
	"github.com/pkg/errors"
)

// Storage ...
type Storage interface {
	GetLastCommandByUserName(ctx context.Context, usename string) (*apitypes.LastUserCommand, error)
	CheckingPigeonWork(ctx context.Context, userN string) (bool, error)
	StopPigeonWork(ctx context.Context, userN string) error
	GetUserByID(ctx context.Context, idUser string) (*apitypes.UserRow, error)
	StartPigeonWork(ctx context.Context, userN string) error
	SetLastComandUser(ctx context.Context, userN string, command string) error
	AddNewUser(ctx context.Context, userN, userID, chatID string) (*apitypes.UserRow, error)
	AddNewCode(ctx context.Context, codeN string) (*apitypes.CodeRow, error)
	AddRefUserCode(ctx context.Context, codeR string, userIDR string) (*apitypes.RefUserCode, error)
	UpdateRefUserCode(ctx context.Context, codeR string, userR string) (*apitypes.RefUserCode, error)
	DeleteLastCommand(ctx context.Context, userId string, command string) error
	GetRefUserCodeByUserName(ctx context.Context, userN string) (*apitypes.RefUserCode, error)
}

// BotSvc ...
type BotSvc struct {
	storage     Storage
	commandsBot types.Commands
}

// NewBotSvc ...
func NewBotSvc(s Storage, commandsBot types.Commands) *BotSvc {
	return &BotSvc{
		storage:     s,
		commandsBot: commandsBot,
	}
}

// ProcessingCommands ...
func (b *BotSvc) ProcessingCommands(message *tgbotapi.Message, bot *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)

	verificationFlags, err := b.GetVerificationFlags(message)
	if err != nil {
		return errors.Wrap(err, "verificationFlags failed")
	}

	if verificationFlags.Verification && verificationFlags.FreshLastCommand {
		err = b.changingCodeWord(message, bot, msg, verificationFlags.LastCommand.Command)
		if err != nil {
			return errors.Wrap(err, "verificationFreshLastCommand failed")
		}
	} else if verificationFlags.Verification {

		err := b.verificationWorkBot(message, bot, msg)
		if err != nil {
			return errors.Wrap(err, "verificationWorkBot failed")
		}

	} else {
		err = b.offerRregister(message, bot, msg)
		if err != nil {
			return errors.Wrap(err, "offerRregister failed")
		}
	}

	return nil
}

func (b *BotSvc) changingCodeWord(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig, command string) error {
	if command == b.commandsBot.AddNameBot {
		err := b.commandAddNameBot(message, bot, msg)
		if err != nil {
			return errors.Wrap(err, "commandAddNameBot failed")
		}

	} else if command == b.commandsBot.EditNameBot {
		err := b.commandEditNameBot(message, bot, msg)
		if err != nil {
			return errors.Wrap(err, "commandEditNameBot failed")
		}
	}
	return nil
}

func (b *BotSvc) commandAddNameBot(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	result, err := b.addUserCode(bot, message)

	if err != nil {
		return errors.Wrap(err, "Bot name creation failed")
	}

	if result {
		err = b.buttonsStartBotWork(msg.BaseChat.ChatID, bot)
		if err != nil {
			return errors.Wrap(err, "buttonsStartBotWork failed")
		}
	} else {
		err = b.buttonAddNameBot(message, bot)

		if err != nil {
			return errors.Wrap(err, "Adding name bot failed")
		}
	}
	return nil
}

func (b *BotSvc) commandEditNameBot(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	result, err := b.editUserCode(bot, message)
	if err != nil {
		return errors.Wrap(err, "Updating user code failed")
	}

	if result {
		err = b.buttonsStartBotWork(msg.BaseChat.ChatID, bot)

		if err != nil {
			return errors.Wrap(err, "Starting work failed")
		}
	} else {
		err = b.buttonAddNameBot(message, bot)

		if err != nil {
			return errors.Wrap(err, "Adding name bot failed")
		}
	}
	return nil
}

func (b *BotSvc) botDoNotWork(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	if message.Text == emoji.Sprint(b.commandsBot.StartBot) {
		err := b.storage.StartPigeonWork(context.Background(), message.From.UserName)
		if err != nil {
			return errors.Wrap(err, "StartPigeonWork failed")
		}

		err = b.buttonsStopBotWork(msg.BaseChat.ChatID, bot)

		if err != nil {
			return errors.Wrap(err, "buttonsStopBotWork failed")
		}
	} else if message.Text == emoji.Sprint(b.commandsBot.EditCode) {
		err := b.buttonEditNameBot(message, bot)

		if err != nil {
			return errors.Wrap(err, "buttonEditNameBot failed")
		}
	}
	return nil
}

func (b *BotSvc) stopBotWorking(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	err := b.storage.StopPigeonWork(context.Background(), message.From.UserName)
	if err != nil {
		return errors.Wrap(err, "StopPigeonWork failed")
	}

	err = b.buttonsStartBotWork(msg.BaseChat.ChatID, bot)
	if err != nil {
		return errors.Wrap(err, "buttonsStartBotWork failed")
	}

	return nil
}

func (b *BotSvc) offerRregister(message *tgbotapi.Message, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) error {
	command := emoji.Sprint(b.commandsBot.StartBot)
	if message.Text == command {
		err := b.startRegisterUser(msg.BaseChat.ChatID, bot, message)

		if err != nil {
			return errors.Wrap(err, "startRegisterUser failed")
		}
	} else {
		err := b.buttonStart(msg.BaseChat.ChatID, bot)
		if err != nil {
			return errors.Wrap(err, "buttonStart failed")
		}
	}
	return nil
}

func (b *BotSvc) startRegisterUser(chatID int64, bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	if err := b.buttonAddedNewUser(message, bot); err != nil {
		return errors.Wrap(err, "buttonAddedNewUser failed")
	}

	// продолжение работы с юзером, предложить создать кодовое слово
	if err := b.buttonAddNameBot(message, bot); err != nil {
		log.Printf(err.Error())
	}

	return nil
}

func (b *BotSvc) addUserCode(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (bool, error) {

	var rune = utf8.RuneCountInString(message.Text)
	if len(message.Text) >= 6 && rune >= 6 {
		refUP, err := b.saveRefUserCode(message)
		if err != nil {
			return false, errors.Wrap(err, "Saving the bot name  failed")
		}

		if refUP == false {
			log.Printf("Bot name not saved.")
		}

		if refUP == true {
			err = b.storage.DeleteLastCommand(context.Background(), strconv.Itoa(message.From.ID), b.commandsBot.AddNameBot)
			if err != nil {
				return false, errors.Wrap(err, "Delete last command name  failed")
			}
		}
		return refUP, nil
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Слишком короткое имя.")
	_, err := bot.Send(msg)
	if err != nil {
		return false, errors.Wrap(err, "Sending the message failed")
	}

	return false, nil
}

func (b *BotSvc) editUserCode(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (bool, error) {

	var rune = utf8.RuneCountInString(message.Text)
	if len(message.Text) >= 6 && rune >= 6 {
		refUP, err := b.updateRefUserCode(message)
		if err != nil {
			return false, errors.Wrap(err, "Editing  the bot name failed")
		}
		if refUP == false {
			log.Printf("Bot name not updated.")
		}

		if refUP == true {
			err = b.storage.DeleteLastCommand(context.Background(), strconv.Itoa(message.From.ID), b.commandsBot.EditNameBot)
			if err != nil {
				return false, errors.Wrap(err, "Delete last command failed")
			}
		}
		return refUP, nil
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Слишком короткое имя.")
	_, err := bot.Send(msg)
	if err != nil {
		return false, errors.Wrap(err, "Sending the message failed")
	}

	return false, nil
}

func (b *BotSvc) saveNewUser(msg *tgbotapi.Message) (bool, error) {

	userInit, err := b.storage.AddNewUser(context.Background(),
		msg.From.UserName,
		strconv.Itoa(msg.From.ID),
		strconv.FormatInt(msg.Chat.ID, 10))
	if err != nil {
		return false, errors.Wrap(err, "Adding new user failed")
	}

	if userInit == nil {
		return false, nil
	}

	return true, nil
}

func (b *BotSvc) saveRefUserCode(msg *tgbotapi.Message) (bool, error) {
	botName, err := b.storage.AddNewCode(context.Background(), msg.Text)

	if err != nil {
		return false, errors.Wrap(err, "Adding new code failed")
	}
	if botName == nil {
		return false, nil
	}

	addRefUserCode, err := b.storage.AddRefUserCode(context.Background(), botName.Code, strconv.Itoa(msg.From.ID))

	if err != nil {
		return false, errors.Wrap(err, "Adding RefUserCode failed")
	}
	if addRefUserCode == nil {
		return false, nil
	}

	return true, nil
}

func (b *BotSvc) updateRefUserCode(msg *tgbotapi.Message) (bool, error) {
	update, err := b.storage.UpdateRefUserCode(context.Background(), msg.Text, msg.From.UserName)

	if err != nil {
		return false, errors.Wrap(err, "Updating RefUserCode failed")
	}
	if update == nil {
		return false, nil
	}

	return true, nil
}
