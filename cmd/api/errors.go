package main

import "net/http"


func (app *application) faliedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string){
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}