package fun

import (
	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/modules"
	"math/rand"
	"regexp"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func throwCoin(r *irc.Message) {
	r.Action("kicks a coin...")

	var result string
	if rand.Float32() < 0.5 {
		result = "Got Head!"
	} else {
		result = "Got Tail!"
	}

	if result != "" {
		r.Respond(result)
	}
}

func InitCoin(mod *modules.Module) {
	mod.AddResponse(regexp.MustCompile("^((throw|kick) )?coin$"), func(r *modules.Response) {
		throwCoin(r.Message)
	}, nil)
	mod.AddCommand("coin", func(r *modules.Command) {
		throwCoin(r.Message)
	}, nil)
}
