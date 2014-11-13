package fun

import (
	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/modules"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func diceRoll(dice, faces int) string {
	sum := 0.0
	results := make([]string, dice)

	for i := 0; i < dice; i++ {
		roll := math.Floor(rand.Float64()*float64(faces)) + 1
		results[i] = strconv.FormatFloat(roll, 'f', -1, 64)
		sum += roll
	}

	if dice > 1 {
		return strings.Join(results, ", ") + " | SUM = " + strconv.FormatFloat(sum, 'f', -1, 64)
	} else {
		return strings.Join(results, ", ")
	}
}

func throwDice(r *irc.Message) {
	r.Action("kicks the dice to you...")

	result := diceRoll(1, 6)
	if result != "" {
		r.Mention("you rolled: " + result)
	}
}

func InitDice(mod *modules.Module) {
	mod.AddResponse(regexp.MustCompile("^((throw|kick|roll) )?dice"), func(r *modules.Response) {
		throwDice(r.Message)
	}, nil)
	mod.AddCommand("dice", func(r *modules.Command) {
		throwDice(r.Message)
	}, nil)
}
