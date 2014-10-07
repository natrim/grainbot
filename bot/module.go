package bot

type Module interface {
	Activate(*Connection)
	Deactivate(*Connection)
}

type EasyModule struct {
	Init func(*Connection)
	Halt func(*Connection)
}

func (m *EasyModule) Activate(conn *Connection) {
	if m.Init != nil {
		m.Init(conn)
	}
}

func (m *EasyModule) Deactivate(conn *Connection) {
	if m.Halt != nil {
		m.Halt(conn)
	}
}
