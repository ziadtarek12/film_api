package models

import (
	"errors"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var ErrRecordNotFound = errors.New("record doesn't exist")
var ErrEditConflict = errors.New("edit conflict")

type Models struct {
	Films       FilmModel
	Users       UserModel
	Tokens      TokenModel
	Permissions PermissionModel
	Watchlist   WatchlistModel
}

func New(driver neo4j.DriverWithContext) Models {
	return Models{
		Films:       FilmModel{Driver: driver},
		Users:       UserModel{Driver: driver},
		Tokens:      TokenModel{Driver: driver},
		Permissions: PermissionModel{Driver: driver},
		Watchlist:   WatchlistModel{Driver: driver},
	}
}
