package main

import (
	"github.com/natrim/grainbot/bot"

	//put your modules here
	_ "github.com/natrim/grainbot/modules/autojoin"
)

func main() {
	bot.Run() //blocks
}
