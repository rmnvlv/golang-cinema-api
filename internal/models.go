package internal

import "time"

type Movie struct {
	id          int64
	name        string
	description string
	date        time.Time
	rate        int
	actors      []Actor
}

type Actor struct {
	id     int64
	name   string
	gender string
	birth  time.Time
}
