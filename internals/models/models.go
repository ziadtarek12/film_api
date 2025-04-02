package models

import (
	"database/sql"
	"errors"
)

var ErrRecordNotFound = errors.New("record doesn't exist")
var ErrEditConflict = errors.New("edit conflict")

type Models struct {
	Films FilmModel

	Users UserModel

	Tokens TokenModel

	Permissions PermissionModel
}

func New(DB *sql.DB) Models {
	return Models{
		Films:       FilmModel{DB: DB},
		Users:       UserModel{DB: DB},
		Tokens:      TokenModel{DB: DB},
		Permissions: PermissionModel{DB: DB},
	}
}
