package validator

import (
	"regexp"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	URLRX   = regexp.MustCompile(`^https?://([\da-z\.-]+\.[a-z\.]{2,6})([\/\w \.-]*)*\.(jpg|jpeg|png|gif|bmp|svg|webp)(\?.*)?$`)
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}

	return false
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func MatchesEmail(email string) bool {
	return EmailRX.MatchString(email)
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)
	for _, val := range values {
		uniqueValues[val] = true

	}

	return len(values) == len(uniqueValues)
}

func MatchesURL(URL string) bool {
	return URLRX.MatchString(URL)
}
