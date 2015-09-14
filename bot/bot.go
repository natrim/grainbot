package bot

import (
	log "github.com/Sirupsen/logrus"
	"github.com/natrim/grainbot/config"
	"github.com/natrim/grainbot/connection"
	"os"
	"os/signal"
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

// GetConfig returns configuration
func (bot *Bot) GetConfig() *config.Configuration {
	return bot.Config
}

// GetConnection returns connection
func (bot *Bot) GetConnection() *connection.Connection {
	return bot.Connection
}

// Run starts the bot and blocks until death of the bot
func (bot *Bot) Run() {
	if err := bot.Connection.ConnectTo(bot.Config.Server); err != nil {
		log.Fatalln(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		bot.Connection.Disconnect()
	}()

	bot.Connection.Wait() //run until death of connection
}
