package system

import (
	"regexp"
	"runtime"
	"syscall"
	"time"

	"github.com/natrim/grainbot/modules"

	human "github.com/dustin/go-humanize"
	"strconv"
)

// precompile the command regexp
var quitreg = regexp.MustCompile("^quit$")
var restartreg = regexp.MustCompile("^restart$")
var joinpartreg = regexp.MustCompile("^(join|part)( #([^ ]*))?$")
var nickreg = regexp.MustCompile("^nick ([^ ]*)$")
var statsreg = regexp.MustCompile("^stats|mem(ory)?|uptime$")

var startTime time.Time

func init() {
	startTime = time.Now()
}

// InitSystem register's dice commands on module load
func InitSystem(mod *modules.Module) {
	owner := &modules.OwnerPermission{}
	mod.AddResponse(quitreg, func(r *modules.Response) {
		r.Respond("okey, " + r.Nick + "! Goodbye everypony!")
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}, owner)
	mod.AddResponse(restartreg, func(r *modules.Response) {
		r.Respond("okey, " + r.Nick + "! Reebooot!")
		syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
	}, owner)
	mod.AddResponse(joinpartreg, func(r *modules.Response) {
		if r.Matches[1] == "join" {
			if len(r.Matches)-1 >= 3 && r.Matches[3] != "" {
				r.Server.Join("#" + r.Matches[3])
				r.Respond("okey, " + r.Nick + "! let'z join " + "#" + r.Matches[3])
			} else {
				r.Mention("tell me where to join!")
			}
		} else if r.Matches[1] == "part" {
			if len(r.Matches)-1 >= 3 && r.Matches[3] != "" {
				r.Respond("okey, " + r.Nick + "! let'z leave " + "#" + r.Matches[3])
				r.Server.Part("#" + r.Matches[3])
			} else if r.Channel != "" {
				r.Respond("okey, " + r.Nick + " leaving from here!")
				r.Server.Part(r.Channel)
			} else {
				r.Mention("tell me from what to leave!")
			}
		}

	}, owner)
	mod.AddResponse(nickreg, func(r *modules.Response) {
		r.Server.Nick(r.Matches[1])
		r.Respond("okey, " + r.Nick + "! Let'z talk as another pony!")
	}, owner)

	mod.AddResponse(statsreg, func(r *modules.Response) {
		mem := &runtime.MemStats{}
		runtime.ReadMemStats(mem)
		r.Respond("PID: " + strconv.Itoa(syscall.Getpid()) + ", last (re)start: " + human.Time(startTime) + ", sys memory: " + human.Bytes(mem.Sys))
	}, owner)
}
