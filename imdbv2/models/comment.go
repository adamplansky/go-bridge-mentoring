package models

type Comment struct {
	Description string
	User        *User
	Movie       *Movie
}
