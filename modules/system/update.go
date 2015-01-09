package system

import (
	log "github.com/Sirupsen/logrus"
	update "github.com/inconshreveable/go-update"
	"github.com/natrim/grainbot/modules"
	"regexp"
	"syscall"
)

var updatereg = regexp.MustCompile(`^update$`)

func UpdateInit(mod *modules.Module) {
	owner := &modules.OwnerPermission{}
	mod.AddResponse(updatereg, func(r *modules.Response) {

		r.Respond("okey, " + r.Nick + "!")

		err, errRecover := update.New().FromUrl(mod.GetConfig().UpdateUrl)
		if err != nil {
			r.Respondf("Update failed: %v", err)
			if errRecover != nil {
				log.Errorf("Failed to recover bad update: %v!", errRecover)
				log.Fatalf("Program exectuable may be missing!")
			}
			return
		}

		r.Respond("update done!")

		//restart
		syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
	}, owner)
}
