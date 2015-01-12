package main

import (
	"github.com/natrim/grainbot/bot"
	"runtime"
)

func main() {
	if runtime.NumCPU() >= 2 { //if enough cpu's then use
		runtime.GOMAXPROCS(2) //two of them
	}

	grainbot := bot.NewBot()
	grainbot.Run()
}
