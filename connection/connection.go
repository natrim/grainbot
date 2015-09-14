package connection

import (
	"errors"
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"

	"crypto/tls"
	"strings"
	"time"
)

var (
	err_notConnected     error = errors.New("Not connected to server!")
	err_alreadyConnected error = errors.New("Already connected to server!")
	err_noSocket         error = errors.New("No socket found!")
	err_noServer         error = errors.New("No server found!")
	err_Disconnected     error = errors.New("Disconnected!")
)

// Connection struct
type Connection struct {
	wg sync.WaitGroup //the loop's waiting group

	socket net.Conn //socket

	mu sync.RWMutex //connection lock

	isConnected  bool //the connection should be ready
	isRestarting bool //the connection is restarting from connected socket

	nick            string // the nick we should have
	server          string //server host:port
	serverIsSecured bool   //server use ssl
	currentNick     string //the nick the bot currently posses

	error chan error    //chan for error logging
	write chan string   //chan for writing to socket
	exit  chan struct{} //chan for signaling the loops to stop

	KeepAlive time.Duration // how long to keep connection alive
	Timeout   time.Duration // timeout
	PingFreq  time.Duration // ping every

	lastMessage time.Time     //time of last received message
	badness     time.Duration //antiflood
	lastSent    time.Time     //antiflood
}

// NewConnection returns new connection instance
func NewConnection() *Connection {
	return &Connection{nick: "grainbot"}
}

// Connect will connect to defined server
func (conn *Connection) Connect() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.server == "" {
		return err_noServer
	}

	if !hasPort(conn.server) {
		conn.server = net.JoinHostPort(conn.server, "6697")
	}

	if !conn.isConnected {
		conn.initialise()

		var err error
		if conn.isRestarting {
			if conn.serverIsSecured {
				log.Debugf("Reusing connection to tls://%s", conn.server)
			} else {
				log.Debugf("Reusing connection to tcp://%s", conn.server)
			}
			if conn.socket == nil {
				err = err_noSocket
			}
		} else {
			d := net.Dialer{Timeout: conn.Timeout}
			if conn.serverIsSecured {
				log.Debugf("Connecting to tls://%s", conn.server)
				conn.socket, err = tls.DialWithDialer(&d, "tcp", conn.server, nil)
			} else {
				log.Debugf("Connecting to tcp://%s", conn.server)
				conn.socket, err = d.Dial("tcp", conn.server)
			}
		}

		log.Infof("Connected to %s (%s)", conn.server, conn.socket.RemoteAddr())
		conn.isConnected = true

		conn.write = make(chan string, 10)
		conn.error = make(chan error, 2)
		conn.wg.Add(3)
		go conn.readLoop()
		go conn.writeLoop()
		go conn.pingLoop()

		time.AfterFunc(time.Second, conn.login) //delay by sec

		return err
	}

	return err_alreadyConnected
}

func (conn *Connection) login() {
	/*if len(conn.Password) > 0 {
		conn.SendRawf("PASS %s", irc.Password)
	}*/

	conn.Nick(conn.nick)
	conn.SendRawf("USER %s 0.0.0.0 0.0.0.0 :%s", "grainbot", "GRAIN_BOT_V1")
}

// ConnectTo connects to server with server string: server.example.com[:port]
func (conn *Connection) ConnectTo(server string) error {
	if server == "" {
		return err_noServer
	}
	conn.server = server
	return conn.Connect()
}

// ConnectWith connects to server using defined socket
func (conn *Connection) ConnectWith(sock net.Conn) error {
	if sock == nil {
		return err_noSocket
	}
	conn.socket = sock
	conn.server = sock.RemoteAddr().String()
	conn.isRestarting = true
	return conn.Connect()
}

// Connected returns true if the bot is connected to server
func (conn *Connection) Connected() bool {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	return conn.isConnected
}

// Disconnect from server
func (conn *Connection) Disconnect() {
	close(conn.exit)
	conn.wg.Wait()
	conn.socket.Close()
	conn.socket = nil
	conn.ErrorChan() <- err_Disconnected
}

// Reconnect to a server using the current connection
func (conn *Connection) Reconnect() error {
	conn.isConnected = false
	conn.isRestarting = true
	return conn.Connect()
}

// initialise prepares variables
func (conn *Connection) initialise() {
	if conn.KeepAlive == 0 {
		conn.KeepAlive = 4 * time.Minute
	}
	if conn.Timeout == 0 {
		conn.Timeout = 1 * time.Minute
	}
	if conn.PingFreq == 0 {
		conn.PingFreq = 15 * time.Minute
	}
	if !conn.isRestarting {
		conn.socket = nil
	}
	conn.exit = make(chan struct{})
}

// ErrorChan returns channel for error handling
func (conn *Connection) ErrorChan() chan error {
	return conn.error
}

// copied from http.client
// hasPort returns true if string contains port
func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}
