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
	StartStopBot string
	CreateCode   string
	EditCode     string
	RegisterUser string
	StartPigeon  string
	AddNameBot   string
	EditNameBot  string
}
