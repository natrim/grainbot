package modules

// OwnerPermission - only mine permission
type OwnerPermission struct{}

var ownerNick = ""

// Validate validate's me
func (p *OwnerPermission) Validate(nick, user, host string) bool {
	if nick != ownerNick {
		return false
	}
	return true
}
