package autojoin

import "github.com/natrim/grainbot/bot"

func init() {
	bot.RegisterModule("autojoin", &bot.EasyModule{Init: func(m *bot.EasyModule) {
		m.AddHandler("join on ok", func(event *bot.Event) {
			if event.Command == "001" {
				channels := bot.GetConfig().GetStringSlice("autojoin.channels")
				for _, chn := range channels {
					event.Server.Join(chn)
				}
			}
		})
	},
	})
}
