package autojoin

import (
	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/modules"
)

func Init(mod *modules.Module) {
	mod.AddIrcMessageHandler("join on ok", func(event *irc.Message) {
		if event.Command == "001" {
			channels := mod.GetConfig().GetStringSlice("autojoin.channels")
			for _, chn := range channels {
				event.Server.Join(chn)
			}
		}
	}, nil)
}
