package connection

import (
	"bufio"
	log "github.com/Sirupsen/logrus"
	"strings"
	"time"
)

// Read data from a connection. To be used as a goroutine.
func (irc *Connection) readLoop() {
	defer irc.wg.Done()
	defer log.Debug("end readLoop")

	errChan := irc.ErrorChan()

	log.Debug("start readLoop")

	//helper socket gouroutine
	go func() {
		br := bufio.NewReaderSize(irc.socket, 512)
		for {
			if irc.socket == nil {
				return
			}
			// Set a read deadline based on the combined timeout and ping frequency - We should ALWAYS have received a response from the server within the timeout after our own pings
			if irc.socket != nil {
				irc.socket.SetReadDeadline(time.Now().Add(irc.Timeout + irc.PingFreq))
			}

			msg, err := br.ReadString('\n')

			// We got past our blocking read, so bin timeout
			if irc.socket != nil {
				var zero time.Time
				irc.socket.SetReadDeadline(zero)
			}

			if err != nil {
				errChan <- err
				return
			}

			irc.read <- msg
		}
	}()

	for {
		if irc.socket == nil {
			return
		}
		select {
		case <-irc.exit:
			return
		case msg, ok := <-irc.read:
			if !ok || msg == "" {
				return
			}

			log.Debugf("[RECV] %s", strings.TrimSpace(msg))

			irc.lastMessage = time.Now()
			msg = msg[:len(msg)-2] //Remove \r\n
			event := &Event{Raw: msg, Connection: irc}
			if msg[0] == ':' {
				if i := strings.Index(msg, " "); i > -1 {
					event.Source = msg[1:i]
					msg = msg[i+1:]
				} else {
					log.Infof("Misformed msg from server: %#s", msg)
				}

				if i, j := strings.Index(event.Source, "!"), strings.Index(event.Source, "@"); i > -1 && j > -1 {
					event.Nick = event.Source[0:i]
					event.User = event.Source[i+1 : j]
					event.Host = event.Source[j+1 : len(event.Source)]
				}
			}

			split := strings.SplitN(msg, " :", 2)
			args := strings.Split(split[0], " ")
			event.Code = strings.ToUpper(args[0])
			event.Arguments = args[1:]
			if len(split) > 1 {
				event.Arguments = append(event.Arguments, split[1])
			}

			//TODO: handle events
			if event.Code == "NICK" && irc.currentNick == event.Nick { //itz us changing nick
				irc.currentNick = event.Arguments[0]
			} else if event.Code == "PING" { //reply to ping
				irc.SendRawf("PONG %s", event.Arguments[len(event.Arguments)-1])
			}

		}
	}
}

// Loop to write to a connection. To be used as a goroutine.
func (irc *Connection) writeLoop() {
	defer irc.wg.Done()
	defer log.Debug("end writeLoop")
	errChan := irc.ErrorChan()
	log.Debug("start writeLoop")
	for {
		if irc.socket == nil {
			return
		}
		select {
		case <-irc.exit:
			return
		case b, ok := <-irc.write:
			if !ok || b == "" {
				return
			}

			log.Debugf("[WRITE] %s", strings.TrimSpace(b))

			// Set a write deadline based on the time out
			irc.socket.SetWriteDeadline(time.Now().Add(irc.Timeout))

			_, err := irc.socket.Write([]byte(b))

			// Past blocking write, bin timeout
			var zero time.Time
			irc.socket.SetWriteDeadline(zero)

			if err != nil {
				errChan <- err
				return
			}
		}
	}
}

// Pings the server if we have not received any messages for 5 minutes
// to keep the connection alive. To be used as a goroutine.
func (irc *Connection) pingLoop() {
	defer irc.wg.Done()
	defer log.Debug("end pingLoop")
	ticker := time.NewTicker(1 * time.Minute) // Tick every minute for monitoring
	ticker2 := time.NewTicker(irc.PingFreq)   // Tick at the ping frequency.
	log.Debug("start pingLoop")
	for {
		if irc.socket == nil {
			return
		}
		select {
		case <-ticker.C:
			//Ping if we haven't received anything from the server within the keep alive period
			if time.Since(irc.lastMessage) >= irc.KeepAlive {
				irc.SendRawf("PING %d", time.Now().UnixNano())
			}
		case <-ticker2.C:
			//Ping at the ping frequency
			irc.SendRawf("PING %d", time.Now().UnixNano())
			//Try to recapture nickname if it's not as configured.
			if irc.nick != irc.currentNick {
				irc.SendRawf("NICK %s", irc.nick)
			}
		case <-irc.exit:
			ticker.Stop()
			ticker2.Stop()
			return
		}
	}
}

// Wait for connection end
func (irc *Connection) Wait() {
	log.Debug("Waiting for connection end")
	defer log.Debug("Connection end")
	errChan := irc.ErrorChan()
	for irc.Connected() {
		err := <-errChan
		if !irc.Connected() {
			break
		}
		log.Printf("Error, disconnected: %s", err)
		return

		//TODO: reconnect
		/*for irc.Connected() {
			if err = irc.Reconnect(); err != nil { //FIX: not working it sends the error to error chan
				log.Printf("Error while reconnecting: %s\n", err)
				time.Sleep(5 * time.Second)
			} else {
				break
			}
		}*/
	}
}
