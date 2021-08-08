// Package app contains all logic of the microservice.
package app

// Module contains business logic for user methods.
type Module struct {
	user Repo
	hash Hasher
	file FileSvc
	auth AuthSvc
}

// New build and returns new Module for working with user info.
func New(r Repo, h Hasher, a AuthSvc, f FileSvc) *Module {
	return &Module{
		user: r,
		hash: h,
		file: f,
		auth: a,
	}
}
