package modules

import (
	"errors"
	"github.com/natrim/grainbot/config"
	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/permissions"
)

func NewModule(name string, init func(*Module), halt func(*Module)) *Module {
	return &Module{name: name, Init: init, Halt: halt}
}

// Start run's before module loading - only once per bot live
func Start(conn *irc.Connection, conf *config.Configuration) {
	//put owner nick in permission
	ownerNick = conf.Owner
}

// Stop run's before module unloading - only once per bot live
func Stop(conn *irc.Connection, conf *config.Configuration) {
}

type Module struct {
	Init func(*Module)
	Halt func(*Module)
	name string

	connection *irc.Connection
	config     *config.Configuration

	handlers map[string]chan bool
}

func (m *Module) Initialize(conn *irc.Connection, conf *config.Configuration, name string) {
	m.connection = conn
	m.config = conf
	m.name = name
	m.handlers = make(map[string]chan bool)
}

func (m *Module) Name() string {
	return m.name
}

func (m *Module) GetConfig() *config.Configuration {
	return m.config
}

func (m *Module) GetConnection() *irc.Connection {
	return m.connection
}

func (m *Module) Activate() {
	if m.Init != nil {
		m.Init(m)
	}
}

func (m *Module) Deactivate() {
	for name, kill := range m.handlers {
		kill <- true
		delete(m.handlers, name)
	}

	if m.Halt != nil {
		m.Halt(m)
	}
}

func (m *Module) AddIrcMessageHandler(name string, f func(*irc.Message), permission permissions.Permission) error {
	if _, ok := m.handlers[name]; ok {
		return errors.New("Handler with same name already exist's!")
	}

	m.handlers[name] = m.connection.AddHandler(f, permission)
	return nil
}

func (m *Module) RemoveIrcMessageHandler(name string) error {
	if len(m.handlers) < 0 {
		return errors.New("This module has no handlers!")
	}

	kill, ok := m.handlers[name]
	if !ok {
		return errors.New("This handler is not defined")
	}

	kill <- true
	delete(m.handlers, name)

	return nil
}
