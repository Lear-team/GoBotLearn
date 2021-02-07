package config

// Configuration ...
type Configuration struct {
	BotToken        string
	BindAddr        string
	LogLevel        string
	ConnectPostgres string
}

// Commands ...
type Commands struct {
	StartBot     string
	StopBot      string
	CreateCode   string
	EditCode     string
	RegisterUser string
}
