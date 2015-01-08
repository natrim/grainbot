package system

import (
	"regexp"
	"syscall"

	"github.com/natrim/grainbot/modules"
)

// precompile the command regexp
var quitreg = regexp.MustCompile("^quit$")
var restartreg = regexp.MustCompile("^restart$")
var joinpartreg = regexp.MustCompile("^(join|part) #([^ ]*)$")

// InitSystem register's dice commands on module load
func InitSystem(mod *modules.Module) {
	owner := &modules.OwnerPermission{}
	mod.AddResponse(quitreg, func(r *modules.Response) {
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}, owner)
	mod.AddResponse(restartreg, func(r *modules.Response) {
		syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
	}, owner)
	mod.AddResponse(joinpartreg, func(r *modules.Response) {
		if r.Matches[1] == "join" {
			r.Server.Join("#" + r.Matches[2])
		} else if r.Matches[1] == "part" {
			r.Server.Part("#" + r.Matches[2])
		}
	}, owner)
}
