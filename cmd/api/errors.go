package main

import (
	"net/http"

)


func (app *application) faliedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string){
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request){
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}