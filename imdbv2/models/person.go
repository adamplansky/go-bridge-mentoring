package models

import "time"

type Person struct {
	FullName        string    `json:"Person.full_name,omitempty"`
	Born            time.Time `json:"Person.born,omitempty"`
	Description     string    `json:"Person.description,omitempty"`

	ArtistsMovies   []Movie   `json:"Person.artists_movies,omitempty"`
	DirectorsMovies []Movie   `json:"Person.directors_movies,omitempty"`
	//Photos     []string
	//TitlePhoto string
	//// custom type / enum
	//Role string
}
