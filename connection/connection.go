package connection

import (
	"errors"
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"

	"crypto/tls"
	"fmt"
	"strings"
)

var (
	err_notConnected     error = errors.New("Not connected to server!")
	err_alreadyConnected error = errors.New("Already connected to server!")
	err_noSocket         error = errors.New("No socket found!")
	err_noServer         error = errors.New("No server found!")
)

// Connection struct
type Connection struct {
	sock net.Conn

	mu sync.RWMutex

	isConnected  bool
	isRestarting bool

	server          string
	serverIsSecured bool
	currentNick     string

	exit chan struct{}
	wg   sync.WaitGroup
}

// NewConnection returns new connection instance
func NewConnection() *Connection {
	return &Connection{}
}

// Connect will connect to defined server
func (conn *Connection) Connect() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if !hasPort(conn.server) {
		conn.server = net.JoinHostPort(conn.server, "6697")
	}

	if !conn.isConnected {
		var err error
		if conn.isRestarting {
			if conn.serverIsSecured {
				log.Debugf("Reusing connection to tls://%s", conn.server)
			} else {
				log.Debugf("Reusing connection to tcp://%s", conn.server)
			}
			if conn.sock == nil {
				err = err_noSocket
			}
		} else {
			if conn.serverIsSecured {
				log.Debugf("Connecting to tls://%s", conn.server)
				conn.sock, err = tls.Dial("tcp", conn.server, nil)
			} else {
				log.Debugf("Connecting to tcp://%s", conn.server)
				conn.sock, err = net.Dial("tcp", conn.server)
			}
		}

		log.Infof("Connected to %s (%s)", conn.server, conn.sock.RemoteAddr())

		return err
	}

	return err_alreadyConnected
}

// ConnectTo connects to server with server string: server.example.com[:port]
func (conn *Connection) ConnectTo(server string) error {
	if server == nil {
		return err_noServer
	}
	conn.server = server
	return conn.Connect()
}

// ConnectWith connects to server using defined socket
func (conn *Connection) ConnectWith(sock *net.Conn) error {
	if sock == nil {
		return err_noSocket
	}
	conn.sock = sock
	conn.isRestarting = true
	return conn.Connect()
}

// Connected returns true if the bot is connected to server
func (conn *Connection) Connected() bool {
	conn.mu.RLock()
	defer conn.mu.RUnlock()
	return conn.isConnected
}

// initialise prepares variables
func (conn *Connection) initialise() {
	conn.sock = nil
	conn.exit = make(chan struct{})
}

// GetNick returns current bot nick
func (conn *Connection) GetNick() string {
	return conn.currentNick
}

// copied from http.client
// hasPort returns true if string contains port
func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}
