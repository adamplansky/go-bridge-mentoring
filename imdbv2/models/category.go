package models

import "encoding/json"

type Category struct {
	ID    string `json:"uid,omitempty"`
	Title string `json:"Category.title,omitempty"`

	Movies  []Movie    `json:"Category.movies,omitempty"`
}

func (c Category) String() string {
	s, _ := json.MarshalIndent(c, "", "\t")
	return string(s)
}