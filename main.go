package main

import (
	"github.com/natrim/grainbot/modules"
	"github.com/natrim/grainbot/modules/autojoin"
)

var grainbot *Bot

func main() {
	grainbot = NewBot()

	//register modules
	grainbot.RegisterModule(modules.NewModule("autojoin", autojoin.Init, nil))

	grainbot.Run() //blocks
}
