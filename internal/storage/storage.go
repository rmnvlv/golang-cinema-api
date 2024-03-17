package storage

import "errors"

var (
	ErrActorExists = errors.New("actor exists")
	ErrFilmExists  = errors.New("film exists")
)
