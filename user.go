package outlived

import "github.com/bobg/aesite"

type User struct {
	aesite.User
	Born     Date
	Active   bool
	TZName   string
	TZOffset int // deprecated
	TZSector int // see function TZSector
}

func (u *User) GetUser() *aesite.User {
	return &u.User
}

func (u *User) SetUser(au *aesite.User) {
	u.User = *au
}

// TZSector reduces a timezone offset (in seconds east of UTC)
// to an index in the range [0 .. 12]
func TZSector(tzoffset int) int {
	return (tzoffset + 36000) / 7200
}
