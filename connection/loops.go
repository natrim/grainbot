package connection

import (
	"bufio"
	"strings"
	"time"
)

// Read data from a connection. To be used as a goroutine.
func (irc *Connection) readLoop() {
	defer irc.Done()
	br := bufio.NewReaderSize(irc.socket, 512)

	errChan := irc.ErrorChan()

	for {
		select {
		case <-irc.exit:
			return
		default:
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
				break
			}

			irc.Log.Debugf("[RECV] %s\n", strings.TrimSpace(msg))

			irc.lastMessage = time.Now()
			msg = msg[:len(msg)-2] //Remove \r\n
			event := &Event{Raw: msg, Connection: irc}
			if msg[0] == ':' {
				if i := strings.Index(msg, " "); i > -1 {
					event.Source = msg[1:i]
					msg = msg[i+1:]
				} else {
					irc.Log.Infof("Misformed msg from server: %#s\n", msg)
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

			if event.Code == "NICK" && irc.currentNick == event.Nick { //itz us changing nick
				irc.currentNick = event.Arguments[0]
			}
			//TODO: handle event
		}
	}
}

// Loop to write to a connection. To be used as a goroutine.
func (irc *Connection) writeLoop() {
	defer irc.Done()
	errChan := irc.ErrorChan()
	for {
		select {
		case <-irc.exit:
			return
		default:
			b, ok := <-irc.write
			if !ok || b == "" || irc.socket == nil {
				return
			}

			irc.Log.Debugf("[WRITE] %s\n", strings.TrimSpace(b))

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
	defer irc.Done()
	ticker := time.NewTicker(1 * time.Minute) // Tick every minute for monitoring
	ticker2 := time.NewTicker(irc.PingFreq)   // Tick at the ping frequency.
	for {
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

// Main loop to control the connection.
func (irc *Connection) Loop() {
	errChan := irc.ErrorChan()
	for irc.Connected() {
		err := <-errChan
		if !irc.Connected() {
			break
		}
		irc.Log.Printf("Error, disconnected: %s\n", err)
		for irc.Connected() {
			if err = irc.Reconnect(); err != nil {
				irc.Log.Printf("Error while reconnecting: %s\n", err)
				time.Sleep(5 * time.Second)
			} else {
				break
			}
		}
	}
}
