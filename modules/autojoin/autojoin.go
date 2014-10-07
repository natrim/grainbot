package autojoin

import "github.com/natrim/grainbot/bot"

var kill chan bool

/*
//method1 struct and ...
type Autojoin struct {
}

func (mod Autojoin) Activate(connection *bot.Connection) {
	kill = connection.AddHandler(func(event *bot.Event) {
		if event.Command == "001" {
			chans := bot.GetConfig().GetStringSlice("autojoin")
			for _, chn := range chans {
				event.Server.Join(chn)
			}
		}
	})
}

func (mod Autojoin) Deactivate(irc *bot.Connection) {
	if kill != nil {
		kill <- true //kill handler
	}
}
*/
func init() {
	//method1: make struct which has Activate and Deactivate
	//bot.RegisterModule("autojoin", &Autojoin{})
	//or just use EasyModule
	bot.RegisterModule("autojoin", &bot.EasyModule{Init: func(connection *bot.Connection) {
		kill = connection.AddHandler(func(event *bot.Event) {
			if event.Command == "001" {
				channels := bot.GetConfig().GetStringSlice("autojoin.channels")
				for _, chn := range channels {
					event.Server.Join(chn)
				}
			}
		})
	}, Halt: func(_ *bot.Connection) {
		if kill != nil {
			kill <- true //kill handler
		}
	}})
}
