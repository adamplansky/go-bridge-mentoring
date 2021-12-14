package models

type Score struct {
	Score int
	User  *User
	Movie *Movie
}

type User struct {
	Name              string
	AverageMovieScore int
}
