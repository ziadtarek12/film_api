package models

import "database/sql"

type Models struct{
	Films interface{
		Get(id int64) (*Film, error)
		Insert(*Film) error
		Update(*Film) error
		Delete(id int64) error
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