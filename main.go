package main

import (
	"github.com/natrim/grainbot/modules"
	"github.com/natrim/grainbot/modules/autojoin"
	"github.com/natrim/grainbot/modules/fun"
	"github.com/natrim/grainbot/modules/system"
)

var grainbot *Bot

func main() {
	grainbot = NewBot()

	//register modules
	grainbot.RegisterModule(modules.NewModule("system", system.InitSystem, nil))
	grainbot.RegisterModule(modules.NewModule("system-update", system.UpdateInit, nil))
	grainbot.RegisterModule(modules.NewModule("autojoin", autojoin.Init, nil))
	grainbot.RegisterModule(modules.NewModule("coin", fun.InitCoin, nil))
	grainbot.RegisterModule(modules.NewModule("dice", fun.InitDice, nil))

	grainbot.Run() //blocks
}
