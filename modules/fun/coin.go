package fun

import (
	"math/rand"
	"regexp"
	"time"

	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/modules"
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

var coinreg = regexp.MustCompile("^((throw|kick) )?coin( (.*))?$")

// InitCoin register's coin command on module load
func InitCoin(mod *modules.Module) {
	mod.AddResponse(coinreg, func(r *modules.Response) {
		throwCoin(r.Message)
	}, nil)
	mod.AddCommand("coin", func(r *modules.Command) {
		throwCoin(r.Message)
	}, nil)
}
