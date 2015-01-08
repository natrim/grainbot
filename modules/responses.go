package modules

import (
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/natrim/grainbot/irc"
	"github.com/natrim/grainbot/permissions"
)

type Response struct {
	*irc.Message
	Text    string
	Matches []string
}

func (m *Module) AddResponse(reg *regexp.Regexp, f func(*Response), permission permissions.Permission) error {
	name := reg.String()
	wrap := func(m *irc.Message) {
		switch m.Command {
		case "PRIVMSG", "NOTICE":
			nick := m.Server.CurrentNick()
			text := strings.Join(m.Arguments[1:], " ")
			if m.Arguments[0] == nick { //direct privmsg or asked from channel
				if reg.MatchString(strings.Trim(text, " ")) {
					f(&Response{m, text, reg.FindStringSubmatch(text)})
				}
			} else {
				current, err := regexp.Compile("^" + nick + "[ ,;:]")
				if err != nil {
					log.Println("Failed to compile nick regexp: ", err)
				} else if current.MatchString(text) {
					nl := len(nick) + 1
					if len(text) > nl {
						just_text := text[nl:]
						if reg.MatchString(strings.Trim(just_text, " ")) {
							f(&Response{m, text, reg.FindStringSubmatch(just_text)})
						}
					}
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

func (m *Module) RemoveResponse(reg *regexp.Regexp) error {
	name := reg.String()

	if len(m.handlers) < 0 {
		return errors.New("This module has no responses!")
	}

	kill, ok := m.handlers[name]
	if !ok {
		return errors.New("This response is not defined")
	}

	kill <- true
	delete(m.handlers, name)

	return nil
}
