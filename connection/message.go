package connection

// A struct to represent an event.
type Event struct {
	Code       string
	Raw        string
	Nick       string //<nick>
	Host       string //<nick>!<usr>@<host>
	Source     string //<host>
	User       string //<usr>
	Arguments  []string
	Connection *Connection
}

// Retrieve the last message from Event arguments.
// This function  leaves the arguments untouched and
// returns an empty string if there are none.
func (e *Event) Message() string {
	if len(e.Arguments) == 0 {
		return ""
	}
	return e.Arguments[len(e.Arguments)-1]
}
