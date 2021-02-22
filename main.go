package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"GoBotPigeon/app/sqlapi"
	"GoBotPigeon/service"
	"GoBotPigeon/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // here
)

var (
	BotToken        = flag.String("BotToken", "telegram_token", "Path to file keys.json")
	BindAddr        = flag.String("BindAddr", "localhost", "Path to file keys.json")
	LogLevel        = flag.String("LogLevel", "debug", "Path to file keys.json")
	ConnectPostgres = flag.String("ConnectPostgres", "localhost", "Path to file keys.json")
)

func main() {
	flag.Parse()

	commandsBot, err := getCommands()
	if err != nil {
		log.Fatal(err)
	}

	db, err := connectDB(*ConnectPostgres)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	apiStorage := sqlapi.NewAPI(db)
	svc := service.NewBotSvc(apiStorage, commandsBot)

	bot, err := tgbotapi.NewBotAPI(*BotToken)
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

		if err := svc.ProcessingCommands(update.Message, bot); err != nil {
			log.Println("Processing commands: ", err.Error())
		}
	}
}

// connectDB ...
func connectDB(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", databaseURL)
	if err != nil {
		log.Println("sqlx.Open failed with an error: ", err.Error())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Println("DB.Ping failed with an error: ", err.Error())
		return nil, err
	}

	return db, err
}

func getCommands() (types.Commands, error) {
	data, err := ioutil.ReadFile("./config/command.json")
	if err != nil {
		log.Println("Reading the command file ended with an error: ", err.Error())
		return types.Commands{}, err
	}

	cmd := types.Commands{}
	err = json.Unmarshal(data, &cmd)
	if err != nil {
		log.Println("Unmarshal ended with an error: ", err.Error())
		return types.Commands{}, err
	}

	return cmd, nil
}
