package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"filmapi.zeyadtarek.net/internals/validator"
)


func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error{
	js, err := json.Marshal(data)
	if err != nil{
		return err
	}

	js = append(js, '\n')

	for key, value := range headers{
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any){
	env := map[string]any{
		"error": message,
	}

	err := app.writeJSON(w, status, env, nil)
	if err != nil{
		app.errorLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error){
	app.errorLogger.Println(err)
	message := "The server encountred a problem and could not process your request"

	app.errorResponse(w, r, http.StatusInternalServerError, message)

}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request){
	message := "The requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request){
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error){
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		
		default:
			return err
		}
	}

	// Check if there's any remaining data in the request body
	if dec.More() {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) readString(queryString url.Values, key string, defaultValue string) string{
	s := queryString.Get(key)
	if s == ""{
		return defaultValue
	}

	return s
}

func (app *application) readCSV(queryString url.Values, key string, defaultValue []string)[]string{
	csv := queryString.Get(key)
	if csv == ""{
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(queryString url.Values, key string, defaultValue int, v *validator.Validator) int{
	str := queryString.Get(key)
	if str == ""{
		return defaultValue
	}

	integer, err := strconv.Atoi(str)
	if err != nil{
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return integer
}

