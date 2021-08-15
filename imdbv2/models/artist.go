package models

import "time"

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
