package bot

import "errors"

type Module interface {
	Initialize(connection *Connection, name string)
	Activate()
	Deactivate()
}

type EasyModule struct {
	Init       func(*EasyModule)
	Halt       func(*EasyModule)
	name       string
	connection *Connection
}

func (m *EasyModule) Initialize(conn *Connection, name string) {
	m.connection = conn
	m.name = name
}

func (m *EasyModule) Name() string {
	return m.name
}

func (m *EasyModule) Activate() {
	if m.Init != nil {
		m.Init(m)
	}
}

func (m *EasyModule) Deactivate() {
	if _, ok := easyHandlers[m.name]; ok {
		for name, kill := range easyHandlers[m.name] {
			kill <- true
			delete(easyHandlers[m.name], name)
		}
	}

	if m.Halt != nil {
		m.Halt(m)
	}
}

var easyHandlers = make(map[string]map[string]chan bool)

func (m *EasyModule) AddHandler(name string, f func(*Event)) error {
	if _, ok := easyHandlers[m.name]; !ok {
		easyHandlers[m.name] = map[string]chan bool{name: m.connection.AddHandler(f)}
		return nil
	}

	if _, ok := easyHandlers[m.name][name]; ok {
		return errors.New("Handler with same name already exist's!")
	}

	easyHandlers[m.name][name] = m.connection.AddHandler(f)
	return nil
}

func (m *EasyModule) RemoveHandler(name string) error {
	if _, ok := easyHandlers[m.name]; !ok {
		return errors.New("This module has no handlers!")
	}

	kill, ok := easyHandlers[m.name][name]
	if !ok {
		return errors.New("This handler is not defined")
	}

	kill <- true
	delete(easyHandlers[m.name], name)

	return nil
}
