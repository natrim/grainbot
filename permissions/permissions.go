package permissions

import "errors"

type Permission interface {
	Validate(nick, user, host string) bool
}

type Allow struct{}
type Deny struct{}

func (p *Allow) Validate(nick, user, host string) bool {
	return true
}

func (p *Deny) Validate(nick, user, host string) bool {
	return false
}

var permissionList = make(map[string]Permission)

func NewPermission(name string, allowordeny bool) Permission {
	var per Permission
	if allowordeny {
		per = &Allow{}
	} else {
		per = &Deny{}
	}

	permissionList[name] = per

	return per
}

func AddPermission(name string, p Permission) error {
	if _, ok := permissionList[name]; ok {
		return errors.New("Permission with this name already exist's!")
	}

	permissionList[name] = p

	return nil
}

func RemovePermission(name string) error {
	if _, ok := permissionList[name]; !ok {
		return errors.New("Permission with this name does'nt exist!")
	}

	delete(permissionList, name)

	return nil
}

func CheckPermission(name, nick, user, host string) bool {
	if p, ok := permissionList[name]; ok {
		return p.Validate(nick, user, host)
	}

	return false
}
