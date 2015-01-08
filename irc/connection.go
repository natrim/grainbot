package irc

import (
	"github.com/natrim/grainbot/broadcast"

	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type Connection struct {
	Hostname    string //hostname to connect
	Port        int    //port to connect
	Secured     bool   //use ssl connection?
	IsConnected bool   //is server connected?

	restarting   bool //is bot restarting itself?
	reconnecting bool //is bot reconnecting to irc?

	Socket net.Conn //connection socket

	Nickname string //nickname the client will use
	Password string //password used to log on to the server
	Username string //supplied to the server as the "User name""
	RealName string //supplied to the server as "Real name" or "ircname"

	heartbeatInterval float64 //interval, in seconds, to send PING messages for keepalive

	// Communication channels
	write     chan string            // Channel for writing messages to IRC server
	broadcast *broadcast.Broadcaster // Channel like broadcasting of the irc messages
	exit      chan struct{}          // Channel for notifying goroutine stop
	ErrorChan chan error             // Channel for dumping errors

	lastMessage     string    //message received as raw string
	lastMessageTime time.Time //time of last message received
	currentNickname string    //current nick
}

func NewConnection(nick, user, realname string) (irc *Connection) {
	irc = &Connection{
		Nickname:  nick,
		Username:  user,
		RealName:  realname,
		broadcast: broadcast.NewBroadcaster(1024),
	}

	irc.AddHandler(defaultHandlers, nil)

	return irc
}

func (irc *Connection) ConnectTo(socket net.Conn) error {
	if socket != nil {
		irc.Socket = socket
		irc.restarting = true
		return irc.Connect()
	}

	return errors.New("No socket connection found!")
}

func (irc *Connection) Connect() error {
	if !irc.IsConnected {
		var err error

		if irc.restarting {
			if irc.Secured {
				log.Printf("Reusing connection to tls://%s:%d\n", irc.Hostname, irc.Port)
			} else {
				log.Printf("Reusing connection to tcp://%s:%d\n", irc.Hostname, irc.Port)
			}
		} else {
			if irc.Secured {
				log.Printf("Connecting to tls://%s:%d\n", irc.Hostname, irc.Port)
				irc.Socket, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", irc.Hostname, irc.Port), nil)
			} else {
				log.Printf("Connecting to tcp://%s:%d\n", irc.Hostname, irc.Port)
				irc.Socket, err = net.Dial("tcp", fmt.Sprintf("%s:%d", irc.Hostname, irc.Port))
			}
			if err != nil {
				return err
			}
		}

		log.Printf("Connected to %s (%s)\n", irc.Hostname, irc.Socket.RemoteAddr())

		irc.write = make(chan string, 1024)
		irc.exit = make(chan struct{})
		irc.ErrorChan = make(chan error, 2)

		irc.lastMessage = ""
		irc.lastMessageTime = time.Now()
		irc.currentNickname = irc.Nickname
		irc.IsConnected = true

		go irc.readLoop()
		go irc.writeLoop()
		go irc.pingLoop()

		//take care of the inital flush
		irc.postConnect()

		//these two needs to be here because some func before use them
		irc.reconnecting = false
		irc.restarting = false

		return nil
	}

	return errors.New("Already connected!")
}

func (irc *Connection) Disconnect() error {
	if !irc.restarting { //pri restartu je uklid jiz drive
		irc.cleanUp()
	}

	if irc.IsConnected {
		err := irc.Socket.Close()
		irc.Socket = nil
		irc.IsConnected = false

		if !irc.restarting {
			log.Printf("Server disconnected.\n")
		}

		return err
	}

	return errors.New("Not connected!")
}

func (irc *Connection) cleanUp() {
	if irc.IsConnected {
		close(irc.write)
	}
	close(irc.exit)
}

func (irc *Connection) Reconnect() error {
	irc.reconnecting = true
	irc.Disconnect()
	return irc.Connect()
}

func (irc *Connection) Restart() {
	irc.restarting = true
	irc.cleanUp()
}

//send raw irc message
func (irc *Connection) SendRaw(message string) {
	irc.write <- strings.Trim(message, "\r\n") + "\r\n"
}

//send raw irc message formated by string
func (irc *Connection) SendRawf(format string, a ...interface{}) {
	irc.SendRaw(fmt.Sprintf(format, a...))
}

//loops

func (irc *Connection) readLoop() {
	br := bufio.NewReaderSize(irc.Socket, 512)

	for {
		select {
		case <-irc.exit:
			return
		default:
			msg, err := br.ReadString('\n')

			if err != nil {
				if err != io.EOF {
					irc.ErrorChan <- err
				}
				return
			}

			irc.lastMessage = msg
			irc.lastMessageTime = time.Now()
			msg = strings.Trim(msg, "\r\n")

			log.Printf("[RECV]<< %s", msg)

			// Publish on broadcast channel
			irc.broadcast.Write(irc.parseIRCMessage(msg))
		}
	}
}

func (irc *Connection) writeLoop() {
	for {
		select {
		case <-irc.exit:
			return
		case b, ok := <-irc.write:
			if !ok || b == "" || irc.Socket == nil {
				return
			}

			log.Printf("[SEND]>> %s", strings.Trim(b, "\r\n"))

			_, err := irc.Socket.Write([]byte(b))
			if err != nil {
				irc.ErrorChan <- err
				return
			}
		}
	}
}

//Pings the server if we have not recived any messages for 5 minutes
func (irc *Connection) pingLoop() {
	ticker := time.NewTicker(1 * time.Minute)   //Tick every minute.
	ticker2 := time.NewTicker(15 * time.Minute) //Tick every 15 minutes.
	for {
		select {
		case <-irc.exit:
			// Shut down everything
			ticker.Stop()
			ticker2.Stop()
			return
		case <-ticker.C:
			// Ping if we haven't received anything from the server within 4 minutes
			if time.Since(irc.lastMessageTime) >= (4 * time.Minute) {
				irc.SendRawf("PING %d", time.Now().UnixNano())
			}
		case <-ticker2.C:
			// Ping every 15 minutes.
			irc.SendRawf("PING %d", time.Now().UnixNano())

			// Try to recapture nickname if it's not as configured.
			if irc.Nickname != irc.currentNickname {
				irc.currentNickname = irc.Nickname
				irc.SendRawf("NICK %s", irc.Nickname)
			}
		}
	}
}

func (irc *Connection) postConnect() {
	if irc.restarting {
		irc.Nick(irc.Nickname) //try original nick
	} else {
		if len(irc.Password) > 0 {
			irc.SendRawf("PASS %s", irc.Password)
		}

		irc.Nick(irc.Nickname)

		realname := irc.RealName
		if irc.RealName == "" {
			realname = irc.Username
		}

		irc.SendRawf("USER %s 0.0.0.0 0.0.0.0 :%s", irc.Username, realname)
	}
}

//irc commands

func (irc *Connection) Nick(n string) {
	irc.currentNickname = n
	irc.SendRawf("NICK %s", n)
}

func (irc *Connection) CurrentNick() string {
	return irc.currentNickname
}

func (irc *Connection) GetNick() string {
	return irc.currentNickname
}

func (irc *Connection) Quit() {
	irc.SendRaw("QUIT")
}

func (irc *Connection) QuitWithMessage(message string) {
	irc.SendRawf("QUIT :%s", message)
}

func (irc *Connection) Join(channel string) {
	irc.SendRawf("JOIN %s", channel)
}

func (irc *Connection) Part(channel string) {
	irc.SendRawf("PART %s", channel)
}

func (irc *Connection) Notice(target, message string) {
	irc.SendRawf("NOTICE %s :%s", target, message)
}

func (irc *Connection) Noticef(target, format string, a ...interface{}) {
	irc.Notice(target, fmt.Sprintf(format, a...))
}

func (irc *Connection) Privmsg(target, message string) {
	irc.SendRawf("PRIVMSG %s :%s", target, message)
}

func (irc *Connection) Privmsgf(target, format string, a ...interface{}) {
	irc.Privmsg(target, fmt.Sprintf(format, a...))
}

func (irc *Connection) Ctcp(target, message string) {
	irc.SendRawf("PRIVMSG %s :\x01%s\x01", target, message)
}

func (irc *Connection) Ctcpf(target, format string, a ...interface{}) {
	irc.Ctcp(target, fmt.Sprintf(format, a...))
}

func (irc *Connection) Ctcpn(target, message string) {
	irc.SendRawf("NOTICE %s :\x01%s\x01", target, message)
}

func (irc *Connection) Ctcpnf(target, format string, a ...interface{}) {
	irc.Ctcpn(target, fmt.Sprintf(format, a...))
}

func (irc *Connection) Action(target, message string) {
	irc.Ctcp(target, "ACTION "+message)
}

func (irc *Connection) Actionf(target, format string, a ...interface{}) {
	irc.Action(target, fmt.Sprintf(format, a...))
}
