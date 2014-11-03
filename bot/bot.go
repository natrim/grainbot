package bot

import (
	"log"

	"flag"
	"github.com/natrim/grainbot/config"
	"strings"
	"sync"
)

type Bot struct {
	Config     *config.Configuration
	Connection *Connection
	modules    map[string]Module
	mwg        *sync.WaitGroup
	restarting bool
}

var grainbot *Bot
var generateConfig = flag.Bool("config", false, "Generate empty config if not exists?")

func init() {
	grainbot = &Bot{Config: config.NewConfiguration(), Connection: NewConnection("dashy", "grainbot", "Botus Grainus"), modules: make(map[string]Module), mwg: &sync.WaitGroup{}}

	//parse flags
	flag.Parse()
}

func GetConfig() *config.Configuration {
	return grainbot.Config
}

func GetConnection() *Connection {
	return grainbot.Connection
}

func RegisterModule(name string, module Module) {
	lname := strings.ToLower(name)
	if grainbot.modules[lname] == nil {
		grainbot.modules[lname] = module
		module.Initialize(grainbot.Connection, name)
	} else {
		log.Println("Cannot register module \"" + name + "\", module with same name already exists!")
	}
}

func Run() {
	var err error
	log.Printf("GRAINBOT - GRAIN based IRC bot ( pid: %d )\n\n", Getpid())

	//load config
	err = grainbot.Config.Load()
	if err != nil {
		if !*generateConfig {
			log.Fatalln(err)
			return
		}

		log.Println("Generating example config.")
		grainbot.Config = config.ExampleConfig()
	}
	log.Println("Config loaded.")

	//load modules
	log.Println("Loading modules...")
	for modname, module := range grainbot.modules {
		if module != nil {
			grainbot.mwg.Add(1)
			module.Activate()
			log.Println("Module \"" + modname + "\" loaded.")
		}
	}
	log.Println("Modules loaded.")

	//set connection
	if grainbot.Config.HostName != "" {
		grainbot.Connection.Hostname = grainbot.Config.HostName
	} else {
		log.Fatalln("No hostname defined!")
		return
	}
	if grainbot.Config.Port != 0 {
		grainbot.Connection.Port = grainbot.Config.Port
	} else {
		grainbot.Connection.Port = 6667
	}
	grainbot.Connection.Secured = grainbot.Config.SSL

	if grainbot.Config.Nick != "" {
		grainbot.Connection.Nickname = grainbot.Config.Nick
	}

	if grainbot.Config.UserName != "" {
		grainbot.Connection.Username = grainbot.Config.UserName
	}

	if grainbot.Config.RealName != "" {
		grainbot.Connection.RealName = grainbot.Config.RealName
	}

	//connect
	if err := grainbot.Connection.Connect(); err != nil {
		log.Fatalln(err)
		return
	}

	//reconnect on error
	go func() {
		for grainbot.Connection.IsConnected {
			err := <-grainbot.Connection.ErrorChan
			if !grainbot.Connection.IsConnected {
				return
			}
			log.Printf("error: %s\n", err)
			grainbot.Connection.Reconnect()
		}
	}()

	//cekej na signal k ukonceni
	if err := WaitOnSignals(grainbot.Connection.Socket); err != nil {
		log.Fatalln(err)
	}

	//prvni sigusr2 signal neukonci Wait
	//zde se dostanem az pri dalsim sigusr2 nebo sigquit

	//ukonci
	if err := grainbot.Connection.Disconnect(); err != nil {
		log.Fatalln(err)
	}

	//unload modules
	if !grainbot.restarting {
		log.Println("Unloading modules...")
		for modname, module := range grainbot.modules {
			if module != nil {
				module.Deactivate()
				grainbot.mwg.Done()
				log.Println("Module \"" + modname + "\" unloaded.")
			}
		}
		grainbot.mwg.Wait() //wait for closing of all
		log.Println("Modules unloaded.")
	}

	//save config
	if !grainbot.restarting {
		err = grainbot.Config.Save()
		if err != nil {
			log.Println("Config save failed.")
			log.Fatalln(err)
		} else {
			log.Println("Config saved.")
		}
	}

	log.Printf("GRAINBOT ( pid: %d ) TERMINATED\n\n", Getpid())
}

func Halt() {
	grainbot.Connection.Disconnect()
}
