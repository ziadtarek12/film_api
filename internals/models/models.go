package models

import (
	"database/sql"
	"errors"
)
var ErrRecordNotFound = errors.New("record doesn't exist")
var ErrEditConflict = errors.New("edit conflict")

type Models struct{
	Films interface{
		Get(id int64) (*Film, error)
		Insert(*Film) error
		Update(*Film) error
		Delete(id int64) error
		Lock(id int64) error
	}

	Directors interface{
		Get(id int64) (*Film, error)
		Insert(*Film) error
		Update(*Film) error
		Delete(id int64) error
	}
}

func New(DB *sql.DB) Models{
	return Models{
		Films: FilmModel{DB: DB},
	}
}