package validation

import (
	"net/mail"
	"strings"
)

type Errors map[string]string

func (e Errors) Add(field string, message string) {
	if _, exists := e[field]; exists {
		return
	}

	e[field] = message
}

func (e Errors) HasAny() bool {
	return len(e) > 0
}

func ValidateRegisterAuth(name string, email string, password string) Errors {
	errors := Errors{}

	if strings.TrimSpace(name) == "" {
		errors.Add("name", "name is required")
	}

	validateEmail(errors, email)
	validatePassword(errors, password)

	return errors
}

func ValidateLoginAuth(email string, password string) Errors {
	errors := Errors{}

	validateEmail(errors, email)
	validatePassword(errors, password)

	return errors
}

func validateEmail(errors Errors, email string) {
	email = strings.TrimSpace(email)
	if email == "" {
		errors.Add("email", "email is required")
		return
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		errors.Add("email", "email must be valid")
	}
}

func validatePassword(errors Errors, password string) {
	if password == "" {
		errors.Add("password", "password is required")
		return
	}

	if len(password) < 8 {
		errors.Add("password", "password must be at least 8 characters")
	}
}
