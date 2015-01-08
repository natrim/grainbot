package fun

import (
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/modules"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// the dice rolling function
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
	}

	return strings.Join(results, ", ")
}

// handle the irc message
func throwDice(r *irc.Message, dice, faces int) {
	r.Action("kicks the dice to you...")

	result := diceRoll(dice, faces)
	if result != "" {
		r.Mention("you rolled: " + result)
	}
}

// precompile the command regexp
var dicereg = regexp.MustCompile("^((throw|kick|roll) )?dice( ([0-9]+d[0-9]+))?$")
var dicesubreg = regexp.MustCompile("dice ([0-9]+d[0-9]+)")

// InitDice register's dice commands on module load
func InitDice(mod *modules.Module) {
	mod.AddResponse(dicereg, func(r *modules.Response) {
		if len(r.Matches)-1 >= 4 && r.Matches[4] != "" {
			dice := strings.Split(r.Matches[4], "d")
			dices, _ := strconv.Atoi(dice[0])
			faces, _ := strconv.Atoi(dice[1])
			throwDice(r.Message, dices, faces)
		} else {
			throwDice(r.Message, 1, 6)
		}
	}, nil)
	mod.AddCommand("dice", func(r *modules.Command) {
		matches := dicesubreg.FindStringSubmatch(r.Text)
		if len(matches)-1 >= 1 && matches[1] != "" {
			dice := strings.Split(matches[1], "d")
			dices, _ := strconv.Atoi(dice[0])
			faces, _ := strconv.Atoi(dice[1])
			throwDice(r.Message, dices, faces)
		} else {
			throwDice(r.Message, 1, 6)
		}
	}, nil)
}
