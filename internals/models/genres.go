package models

import (
	"encoding/json"
)

type Genre struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (g Genre) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.Name)
}
