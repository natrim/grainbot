package main

import (
	"flag"
	"log"
	"strings"
	"sync"
	"time"

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

func init() {
	flag.Parse()
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
		log.Println("Cannot register module \"" + name + "\", module with same name already exists!")
	}
}

// Run spin's and block's - i mean it runs the bot
func (b *Bot) Run() {
	var err error
	log.Printf("GRAINBOT - GRAIN based IRC bot ( pid: %d )\n\n", Getpid())

	//first try to find saved socket
	socket, err := findSocket()
	if err == nil { //ok we have socket so kill parent first
		if err := killParentAfterRestart(); err != nil {
			log.Fatalln(err)
			return
		}
	}

	//load config
	err = b.Config.Load()
	if err != nil {
		if !*generateConfig {
			log.Fatalln(err)
			return
		}

		log.Println("Generating example config.")
		b.Config = config.ExampleConfig()
	}
	log.Println("Config loaded.")

	//load modules
	log.Println("Loading modules...")
	for modname, mod := range b.modules {
		if mod != nil {
			b.mwg.Add(1)
			mod.Activate()
			log.Println("Module \"" + modname + "\" loaded.")
		}
	}
	log.Println("Modules loaded.")

	//set connection
	if b.Config.HostName != "" {
		b.Connection.Hostname = b.Config.HostName
	} else {
		log.Fatalln("No hostname defined!")
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
			log.Fatalln(err)
			return
		}
	} else {
		//else make new connection
		if err := b.Connection.Connect(); err != nil {
			log.Fatalln(err)
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
			log.Printf("error: %s\n", err)
			log.Println("Reconnecting in 10 seconds...")
			time.Sleep(10 * time.Second)

			err = b.Connection.Reconnect()
			if err != nil {
				log.Printf("error: %s\n", err)
			}
		}
	}()

	//cekej na signal k ukonceni
	if err := b.WaitOnSignals(b.Connection.Socket); err != nil {
		log.Fatalln(err)
	}

	//prvni sigusr2 signal neukonci Wait
	//zde se dostanem az pri dalsim sigusr2 nebo sigquit

	//ukonci
	if err := b.Connection.Disconnect(); err != nil {
		log.Fatalln(err)
	}

	//unload modules
	if !b.restarting {
		log.Println("Unloading modules...")
		for modname, mod := range b.modules {
			if mod != nil {
				mod.Deactivate()
				b.mwg.Done()
				log.Println("Module \"" + modname + "\" unloaded.")
			}
		}
		b.mwg.Wait() //wait for closing of all
		log.Println("Modules unloaded.")
	}

	//save config
	if !b.restarting {
		err = b.Config.Save()
		if err != nil {
			log.Println("Config save failed.")
			log.Fatalln(err)
		} else {
			log.Println("Config saved.")
		}
	}

	log.Printf("GRAINBOT ( pid: %d ) TERMINATED\n\n", Getpid())
}

func (b *Bot) beforeFork() error {
	log.Printf("GRAINBOT ( pid: %d ) RESTARTING\n\n", Getpid())

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
		log.Println("Config save failed.")
		log.Fatalln(err)
	}

	return err
}
