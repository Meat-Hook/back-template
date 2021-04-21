// Package app contains all logic of the microservice.
package app

// Module contains business logic for user methods.
type Module struct {
	user Repo
	hash Hasher
	auth Auth
}

// New build and returns new Module for working with user info.
func New(r Repo, h Hasher, a Auth) *Module {
	return &Module{
		user: r,
		hash: h,
		auth: a,
	}
}
