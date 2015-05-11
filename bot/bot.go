package bot

import (
	"github.com/natrim/grainbot/config"
	"github.com/natrim/grainbot/connection"
)

// Bot struct
type Bot struct {
	Connection *connection.Connection
	Config     *config.Configuration
}

// NewBot returns bot instance
func NewBot() *Bot {
	return &Bot{Connection: connection.NewConnection(), Config: config.NewConfiguration()}
}

// Run starts the bot and blocks until death of the bot
func (bot *Bot) Run() {
	if err := bot.Connection.Connect(); err != nil {
		bot.Connection.Log.Fatalln(err)
		return
	}

	bot.Connection.Loop() //run until death of connection
}
