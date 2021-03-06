package irc

import (
	"github.com/natrim/grainbot/broadcast"

	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
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

	wg sync.WaitGroup //wait group for loops

	lastMessage     string    //message received as raw string
	lastMessageTime time.Time //time of last message received

	// Internal counters for flood protection
	badness  time.Duration
	lastsent time.Time

	currentNickname string //current nick
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
				log.Debugf("Reusing connection to tls://%s:%d", irc.Hostname, irc.Port)
			} else {
				log.Debugf("Reusing connection to tcp://%s:%d", irc.Hostname, irc.Port)
			}
		} else {
			if irc.Secured {
				log.Debugf("Connecting to tls://%s:%d", irc.Hostname, irc.Port)
				irc.Socket, err = tls.Dial("tcp", fmt.Sprintf("%s:%d", irc.Hostname, irc.Port), nil)
			} else {
				log.Debugf("Connecting to tcp://%s:%d", irc.Hostname, irc.Port)
				irc.Socket, err = net.Dial("tcp", fmt.Sprintf("%s:%d", irc.Hostname, irc.Port))
			}
			if err != nil {
				return err
			}
		}

		log.Infof("Connected to %s (%s)", irc.Hostname, irc.Socket.RemoteAddr())

		irc.write = make(chan string, 1024)
		irc.exit = make(chan struct{})
		irc.ErrorChan = make(chan error, 2)

		irc.lastMessage = ""
		irc.lastMessageTime = time.Now()
		irc.lastsent = time.Now()
		irc.currentNickname = irc.Nickname
		irc.IsConnected = true

		irc.wg.Add(3)
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
			log.Info("Server disconnected.")
		}

		irc.wg.Wait() //wait for loop's end's

		return err
	}

	return errors.New("Not connected!")
}

func (irc *Connection) cleanUp() {
	close(irc.exit)
	if irc.IsConnected {
		close(irc.write)
	}
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
	defer irc.wg.Done()
	br := bufio.NewReaderSize(irc.Socket, 512)
	for {
		select {
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

			log.Debugf("[RECV]<< %s", msg)

			// Publish on broadcast channel
			irc.broadcast.Write(irc.parseIRCMessage(msg))
		case <-irc.exit:
			return
		}
	}
}

func (irc *Connection) writeLoop() {
	defer irc.wg.Done()
	for {
		select {
		case b, ok := <-irc.write:
			if !ok || b == "" || irc.Socket == nil {
				return
			}

			if t := irc.rateLimit(len(b)); t != 0 {
				// sleep for the current line's time value before sending it
				log.Infof("Message flood! Sleeping for %.2f secs.", t.Seconds())
				<-time.After(t)
			}

			log.Debugf("[SEND]>> %s", strings.Trim(b, "\r\n"))

			_, err := irc.Socket.Write([]byte(b))
			if err != nil {
				irc.ErrorChan <- err
				return
			}
		case <-irc.exit:
			return
		}
	}
}

//Pings the server if we have not recived any messages for 5 minutes
func (irc *Connection) pingLoop() {
	defer irc.wg.Done()
	ticker1 := time.NewTicker(1 * time.Minute)   //Tick every minute.
	ticker15 := time.NewTicker(15 * time.Minute) //Tick every 15 minutes.
	ticker60 := time.NewTicker(60 * time.Minute) //Tick every 60 minutes.
	for {
		select {
		case <-ticker1.C:
			// Ping if we haven't received anything from the server within 4 minutes
			if time.Since(irc.lastMessageTime) >= (4 * time.Minute) {
				irc.SendRawf("PING %d", time.Now().UnixNano())
			}
		case <-ticker15.C:
			// Ping every 15 minutes.
			irc.SendRawf("PING %d", time.Now().UnixNano())
		case <-ticker60.C:
			// Try to recapture nickname if it's not as configured.
			if irc.Nickname != irc.currentNickname {
				irc.currentNickname = irc.Nickname
				irc.SendRawf("NICK %s", irc.Nickname)
			}
		case <-irc.exit:
			// Shut down everything
			ticker1.Stop()
			ticker15.Stop()
			ticker60.Stop()
			return
		}
	}
}

// Implement Hybrid's flood control algorithm to rate-limit outgoing lines.
func (irc *Connection) rateLimit(chars int) time.Duration {
	// Hybrid's algorithm allows for 2 seconds per line and an additional
	// 1/120 of a second per character on that line.
	linetime := 2*time.Second + time.Duration(chars)*time.Second/120
	elapsed := time.Now().Sub(irc.lastsent)
	if irc.badness += linetime - elapsed; irc.badness < 0 {
		// negative badness times are badness...
		irc.badness = 0
	}
	irc.lastsent = time.Now()
	// If we've sent more than 10 second's worth of lines according to the
	// calculation above, then we're at risk of "Excess Flood".
	if irc.badness > 10*time.Second {
		return linetime
	}
	return 0
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
