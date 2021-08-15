package models

import "time"

type Movie struct {
	Title       string
	ReleaseDate time.Time
	Categories  []Category
	Duration    time.Duration
	Score       int
	Description string
	Image       []byte
	Directors   []Participator
	Keywords    []string
	Comments    []string
}
