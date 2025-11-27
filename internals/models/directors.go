package models

import (
	"encoding/json"
)

type Director struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (d Director) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Name)
}
