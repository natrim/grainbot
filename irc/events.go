package irc

import (
	"github.com/natrim/grainbot/permissions"
	"log"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Raw              string // Raw message string
	Prefix           string
	Command          string
	Arguments        []string
	Server           *Connection
	Nick, User, Host string
}

func (m *Message) reply(message string) {
	m.Server.Privmsg(m.Nick, message)
}

//raw irc string parsing
func (irc *Connection) parseIRCMessage(msg string) *Message {
	// http://twistedmatrix.com/trac/browser/trunk/twisted/words/protocols/irc.py#54
	prefix := ""
	trailing := ""
	command := ""
	nick := ""
	user := ""
	host := ""
	args := []string{}
	s := msg

	if s[0] == ':' {
		splits := strings.SplitN(s[1:], " ", 2)
		prefix, s = splits[0], splits[1]
	}

	if strings.Contains(s, " :") {
		splits := strings.SplitN(s, " :", 2)
		s, trailing = splits[0], splits[1]
		args = strings.Fields(s)
		args = append(args, trailing)
	} else {
		args = strings.Fields(s)
	}
	command, args = args[0], args[1:]

	if i, j := strings.Index(prefix, "!"), strings.Index(prefix, "@"); i > -1 && j > -1 {
		nick = prefix[0:i]
		user = prefix[i+1 : j]
		host = prefix[j+1 : len(prefix)]
	}

	return &Message{
		Raw:       msg,
		Prefix:    prefix,
		Command:   command,
		Arguments: args,
		Server:    irc,
		Nick:      nick,
		User:      user,
		Host:      host,
	}
}

func (irc *Connection) AddHandler(f func(*Message), permission permissions.Permission) chan bool {
	messages := irc.broadcast.Listen(1024)
	killchan := make(chan bool)
	go func() {
		for {
			select {
			case k := <-killchan:
				if k {
					return
				}
			case e := <-messages:
				event := e.(*Message)
				if permission != nil {
					if ok := permission.Validate(event.Nick, event.User, event.Host); ok {
						f(event)
					}
				} else {
					f(event)
				}
			}
		}
	}()
	return killchan
}

func defaultHandlers(event *Message) {
	irc := event.Server

	switch event.Command {
	case "PING":
		irc.SendRawf("PONG %s", event.Arguments[len(event.Arguments)-1])

	case "433":
		irc.currentNickname = irc.currentNickname + "_"
		irc.Nick(irc.currentNickname)

	case "PONG":
		ns, _ := strconv.ParseInt(event.Arguments[1], 10, 64)
		delta := time.Duration(time.Now().UnixNano() - ns)
		log.Printf("Lag: %v\n", delta)

	case "PRIVMSG", "NOTICE":
		if event.Arguments[0] == irc.currentNickname && len(event.Arguments[1]) > 2 && strings.HasPrefix(event.Arguments[1], "\x01") && strings.HasSuffix(event.Arguments[1], "\x01") {
			ctcp := strings.Trim(event.Arguments[1], "\x01")
			parts := strings.Split(ctcp, " ")

			switch parts[0] {
			case "VERSION":
				irc.Ctcpn(event.Nick, "VERSION rainbot:2:grain")

			case "TIME":
				irc.Ctcpnf(event.Nick, "TIME %s", time.Now().Format(time.RFC1123))

			case "PING":
				irc.Ctcpn(event.Nick, ctcp)

			case "USERINFO":
				irc.Ctcpnf(event.Nick, "USERINFO %s", irc.Username)

			case "CLIENTINFO":
				irc.Ctcpnf(event.Nick, "CLIENTINFO PING VERSION TIME USERINFO CLIENTINFO")
			}
		} else if event.Arguments[0] == irc.currentNickname { //direct messages
			//nothing here for now
		}
	case "001":
		//log.Print(event.Arguments)
	}
}