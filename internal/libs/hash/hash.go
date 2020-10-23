// Package hash will hashing and comparable hash.
package hash

import (
	"golang.org/x/crypto/bcrypt"
)

type (
	// Hasher contains method for hashing and comparable value.
	Hasher struct {
		cost int
	}
	// Option for building Password struct.
	Option func(*Hasher)
)

// Cost option for sets hashing cost.
func Cost(cost int) Option {
	return func(password *Hasher) {
		password.cost = cost
	}
}

// New creates and returns new Hasher.
func New(options ...Option) *Hasher {
	h := &Hasher{cost: bcrypt.DefaultCost}

	for i := range options {
		options[i](h)
	}

	return h
}

// Hashing value and returns bytes.
func (p *Hasher) Hashing(val string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(val), p.cost)
}

// Compare comparable two hash.
func (p *Hasher) Compare(val1 []byte, val2 []byte) bool {
	return bcrypt.CompareHashAndPassword(val1, val2) == nil
}
