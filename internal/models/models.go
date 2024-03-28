package models

import "time"

type Movie struct {
	Id          int64
	Title       string
	Description string
	Date        time.Time
	Rating      int
	Actors      []Actor
}

type Actor struct {
	Id     int64
	Name   string
	Gender string
	Birth  time.Time
	Movies []string
}

type SerchMovieParams struct {
	Sort         bool   `json: "sort`
	SortType     string `json: "type-sort"`
	FragmentType string `json: "type-fragment"`
	Fragments    string `json: "fragments"`
}
