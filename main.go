package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/natrim/grainbot/bot"
)

var debug = flag.Bool("debug", false, "Print debug messages?")
var createConfig = flag.Bool("config", false, "Create new grain config?")

func init() {
	flag.Parse()

	log.SetFormatter(&PrettyLogFormatter{})
	if *debug == true {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	grainbot := bot.NewBot()
	if *createConfig == true {
		log.Debug("Loading example configuration to grain")
		grainbot.GetConfig().LoadExampleConfig()
	}
	log.Debug("Starting grain")
	grainbot.Run()
}
