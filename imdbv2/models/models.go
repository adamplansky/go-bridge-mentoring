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

// Participator = Artist / Director
type Participator struct {
	FirstName  string
	LastName   string
	BirthDate  time.Time
	Biography  string
	Movies     []Movie
	Photos     []string
	TitlePhoto string
	// custom type / enum
	Role string
}

type Category struct {
	Title string
}

type Comment struct {
	Description string
	User        *User
	Movie       *Movie
}

type Score struct {
	Score int
	User  *User
	Movie *Movie
}

type User struct {
	Name              string
	AverageMovieScore int
}
