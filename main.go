package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/natrim/grainbot/bot"
	"runtime"
)

var debug = flag.Bool("debug", false, "Print debug messages?")

func init() {
	flag.Parse()

	log.SetFormatter(&PrettyLogFormatter{})
	if *debug == true {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	if runtime.NumCPU() >= 2 { //if enough cpu then use
		runtime.GOMAXPROCS(2) //two of them
	}

	grainbot := bot.NewBot()
	grainbot.Run()
}
