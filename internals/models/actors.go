package models

import (
	"encoding/json"
)

type Actor struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (a Actor) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}
