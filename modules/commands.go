package modules

import (
	"errors"
	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/permissions"
	"strings"
)

const COMMAND_DELIMITER = "."

type Command struct {
	*irc.Message
	Text string
}

func (m *Module) AddCommand(name string, f func(*Command), permission permissions.Permission) error {
	name = COMMAND_DELIMITER + name
	wrap := func(message *irc.Message) {
		switch message.Command {
		case "PRIVMSG", "NOTICE":
			nick := message.Server.CurrentNick()
			text := strings.Join(message.Arguments[1:], " ")
			if message.Arguments[0] != nick { //dot command from channel
				if strings.HasPrefix(strings.Trim(text, " "), name) {
					f(&Command{message, text})
				}
			}
		}
	}

	if _, ok := m.handlers[name]; ok {
		return errors.New("Response with same regexp already exist's!")
	}

	m.handlers[name] = m.connection.AddHandler(wrap, permission)
	return nil
}

func (m *Module) RemoveCommand(name string) error {
	name = COMMAND_DELIMITER + name

	if len(m.handlers) < 0 {
		return errors.New("This module has no commands!")
	}

	kill, ok := m.handlers[name]
	if !ok {
		return errors.New("This command is not defined")
	}

	kill <- true
	delete(m.handlers, name)

	return nil
}
