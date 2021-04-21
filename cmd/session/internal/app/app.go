package app

// Module contains business logic for user methods.
type Module struct {
	session Repo
	user    Users
	auth    Auth
	id      ID
}

// New build and returns new session module.
func New(r Repo, u Users, a Auth, id ID) *Module {
	return &Module{
		session: r,
		user:    u,
		auth:    a,
		id:      id,
	}
}
