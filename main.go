package main

import (
	"runtime"

	"github.com/natrim/grainbot/modules"
	"github.com/natrim/grainbot/modules/autojoin"
	"github.com/natrim/grainbot/modules/fun"
	"github.com/natrim/grainbot/modules/system"
)

var grainbot *Bot

func main() {
	if runtime.NumCPU() >= 2 { //if enough cpu's then use
		runtime.GOMAXPROCS(2) //two of them
	}

	grainbot = NewBot()

	//register modules
	grainbot.RegisterModule(modules.NewModule("system", system.InitSystem, nil))
	grainbot.RegisterModule(modules.NewModule("autojoin", autojoin.Init, nil))
	grainbot.RegisterModule(modules.NewModule("coin", fun.InitCoin, nil))
	grainbot.RegisterModule(modules.NewModule("dice", fun.InitDice, nil))

	grainbot.Run() //blocks
}
