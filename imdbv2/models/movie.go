package models

import (
	"encoding/json"
	"time"
)

type Movie struct {
	ID          string        `json:"uid,omitempty"`
	Title       string        `json:"Movie.title,omitempty"`
	ReleaseDate time.Time     `json:"Movie.release_date,omitempty"`
	Duration    time.Duration `json:"Movie.duration,omitempty"`
	Score       int           `json:"Movie.score,omitempty"`
	Description string        `json:"Movie.description,omitempty"`

	Categories  []Category    `json:"Movie.categories,omitempty"`

	Directors   []Person      `json:"Movie.directors,omitempty"`
	Artists     []Person      `json:"Movie.artists,omitempty"`

	Image    []byte
	Keywords []string
	Comments []string
}

func (m Movie) String() string {
	s, _ := json.MarshalIndent(m, "", "\t")
	return string(s)
}
