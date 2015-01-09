package main

import (
	"flag"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/natrim/grainbot/config"
	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/modules"
)

// Bot is the main struct with aaall the ponies
type Bot struct {
	Config     *config.Configuration
	Connection *irc.Connection
	modules    map[string]*modules.Module
	mwg        *sync.WaitGroup
	restarting bool
}

var generateConfig = flag.Bool("config", false, "Generate empty config if not exists?")
var debug = flag.Bool("debug", false, "Print debug messages?")

func init() {
	flag.Parse()

	log.SetFormatter(&PrettyFormatter{})
	if *debug == true {
		log.SetLevel(log.DebugLevel)
	}
}

// NewBot create's new Bot instance
func NewBot() *Bot {
	return &Bot{Config: config.NewConfiguration(), Connection: irc.NewConnection("dashy", "grainbot", "Botus Grainus"), modules: make(map[string]*modules.Module), mwg: &sync.WaitGroup{}}
}

// RegisterModule register's module into bot
func (b *Bot) RegisterModule(mod *modules.Module) {
	name := mod.Name()
	lname := strings.ToLower(name)
	if b.modules[lname] == nil {
		b.modules[lname] = mod
		mod.Initialize(b.Connection, b.Config, name)
	} else {
		log.Fatal("Cannot register module \"" + name + "\", module with same name already exists!")
	}
}

// Run spin's and block's - i mean it runs the bot
func (b *Bot) Run() {
	var err error

	//first try to find saved socket - AND kill parent .)
	socket, err := findSocket()
	if err == nil { //ok we have socket so kill parent first
		if err := killParentAfterRestart(); err != nil {
			log.Fatal(err)
			return
		}
	}

	//thingies to do on start
	func() {
		log.Infof("GRAINBOT - GRAIN based IRC bot ( pid: %d )", Getpid())

		//load config
		err = b.Config.Load()
		if err != nil {
			if !*generateConfig {
				log.Fatal(err)
				return
			}

			log.Info("Generating example config.")
			b.Config = config.ExampleConfig()
		}
		log.Info("Config loaded.")

		//Start module thingie
		modules.Start(b.Connection, b.Config)

		//load modules
		log.Debug("Loading modules...")
		for modname, mod := range b.modules {
			if mod != nil {
				b.mwg.Add(1)
				mod.Activate()
				log.Debug("Module \"" + modname + "\" loaded.")
			}
		}
		log.Info("Modules loaded.")
	}()

	//set thingies to do on exit
	defer func() {
		defer log.Infof("GRAINBOT ( pid: %d ) TERMINATED", Getpid())

		//module thingie
		modules.Stop(b.Connection, b.Config)

		//unload modules
		if !b.restarting {
			log.Debug("Unloading modules...")
			for modname, mod := range b.modules {
				if mod != nil {
					mod.Deactivate()
					b.mwg.Done()
					log.Debug("Module \"" + modname + "\" unloaded.")
				}
			}
			b.mwg.Wait() //wait for closing of all
			log.Info("Modules unloaded.")
		}

		//save config
		if !b.restarting {
			err = b.Config.Save()
			if err != nil {
				log.Fatalf("Config save failed. %s", err)
			} else {
				log.Info("Config saved.")
			}
		}
	}()

	//set connection
	if b.Config.HostName != "" {
		b.Connection.Hostname = b.Config.HostName
	} else {
		log.Fatal("No hostname defined!")
		return
	}
	if b.Config.Port != 0 {
		b.Connection.Port = b.Config.Port
	} else {
		b.Connection.Port = 6667
	}
	b.Connection.Secured = b.Config.SSL

	if b.Config.Nick != "" {
		b.Connection.Nickname = b.Config.Nick
	}

	if b.Config.UserName != "" {
		b.Connection.Username = b.Config.UserName
	}

	if b.Config.RealName != "" {
		b.Connection.RealName = b.Config.RealName
	}

	//connect
	if socket != nil {
		if err := b.Connection.ConnectTo(socket); err != nil {
			log.Fatal(err)
			return
		}
	} else {
		//else make new connection
		if err := b.Connection.Connect(); err != nil {
			log.Fatal(err)
			return
		}
	}

	//reconnect on error
	go func() {
		for b.Connection.IsConnected {
			err := <-b.Connection.ErrorChan
			if !b.Connection.IsConnected {
				return
			}
			log.Errorf("error: %s", err)
			log.Info("Reconnecting in 10 seconds...")
			time.Sleep(10 * time.Second)

			err = b.Connection.Reconnect()
			if err != nil {
				log.Errorf("error: %s", err)
			}
		}
	}()

	//cekej na signal k ukonceni
	if err := b.WaitOnSignals(b.Connection.Socket); err != nil {
		log.Fatal(err)
	}

	//prvni sigusr2 signal neukonci Wait
	//zde se dostanem az pri dalsim sigusr2 nebo sigquit

	//ukonci
	if err := b.Connection.Disconnect(); err != nil {
		log.Fatal(err)
	}
}

func (b *Bot) beforeFork() error {
	log.Infof("GRAINBOT ( pid: %d ) RESTARTING", Getpid())

	b.restarting = true
	b.Connection.Restart()

	for _, module := range b.modules {
		if module != nil {
			module.Deactivate()
			b.mwg.Done()
		}
	}
	b.mwg.Wait() //wait for closing of all

	//save config right now
	err := b.Config.Save()
	if err != nil {
		log.Fatalf("Config save failed. %s", err)
	}

	return err
}
