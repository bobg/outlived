package outlived

import "github.com/bobg/aesite"

type User struct {
	aesite.User
	Born Date
}

func (u *User) GetUser() *aesite.User {
	return &u.User
}

func (u *User) SetUser(au *aesite.User) {
	u.User = *au
}
