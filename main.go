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
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	BotToken        = flag.String("BotToken", "---", "bot")
	BindAddr        = flag.String("BindAddr", "8080", "localhost")
	LogLevel        = flag.String("LogLevel", "release", "release")
	ConnectPostgres = flag.String("ConnectPostgres", "user=postgres password=postgres dbname=PIgeonDB sslmode=disable", "connect db")
)

func main() {
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	commandsBot, err := getCommands()
	if err != nil {
		log.Fatal(err)
	}

	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	apiStorage := sqlapi.NewAPI(db)
	svc := service.NewBotSvc(apiStorage, commandsBot)

	bot, err := tgbotapi.NewBotAPI(viper.GetString("BotToken"))
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
func connectDB() (*sqlx.DB, error) {
	bc := viper.GetString("ConnectPostgres")

	db, err := sqlx.Open("postgres", bc)
	if err != nil {
		return nil, errors.Wrap(err, "sqlx.Open failed with an error")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "DB.Ping failed with an error")
	}

	return db, err
}

func getCommands() (types.Commands, error) {
	data, err := ioutil.ReadFile("./config/command.json")
	if err != nil {
		return types.Commands{}, errors.Wrap(err, "Reading the command file ended with an error")
	}

	cmd := types.Commands{}
	err = json.Unmarshal(data, &cmd)
	if err != nil {
		return types.Commands{}, errors.Wrap(err, "Unmarshal ended with an error")
	}

	return cmd, nil
}
