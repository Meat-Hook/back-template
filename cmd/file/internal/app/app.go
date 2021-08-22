package app

// MaxChunkSize max size of file chunk.
const MaxChunkSize = 4096

// Module contains business logic for file methods.
type Module struct {
	file Repo
}

// New build and returns new session module.
func New(r Repo) *Module {
	return &Module{
		file: r,
	}
}
