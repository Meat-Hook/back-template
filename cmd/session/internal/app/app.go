package app

// Module contains business logic for session methods.
type Module struct {
	session Repo
	auth    Auth
	id      ID
}

// New build and returns new session module.
func New(r Repo, a Auth, id ID) *Module {
	return &Module{
		session: r,
		auth:    a,
		id:      id,
	}
}
