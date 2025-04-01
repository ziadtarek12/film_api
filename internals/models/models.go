package models

import (
	"database/sql"
	"errors"
	"time"
)
var ErrRecordNotFound = errors.New("record doesn't exist")
var ErrEditConflict = errors.New("edit conflict")

type Models struct{
	Films interface{
		Get(id int64) (*Film, error)
		Insert(*Film) error
		Update(*Film) error
		Delete(id int64) error
		GetAll(title string, genres []string, actors []string, directors []string,filters Filters) ([]*Film, Metadata,error)

	}

	Users interface{
		GetByEmail(string) (*User, error)
		Insert(*User) error
		Update(*User) error
		GetForToken(string, string) (*User, error)
	}

	Tokens interface {
		New(int64, time.Duration, string) (*Token, error)
		Insert(*Token) error
		DeleteAllForUser(string, int64) error
	}
}

func New(DB *sql.DB) Models{
	return Models{
		Films: FilmModel{DB: DB},
		Users: UserModel{DB: DB},
		Tokens: TokenModel{DB: DB},
	}
}