package validation

import (
	"strconv"
	"strings"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

func ValidateProjectPagination(pageParam string, limitParam string) (int, int, Errors) {
	errors := Errors{}

	page := defaultPage
	if strings.TrimSpace(pageParam) != "" {
		parsedPage, err := strconv.Atoi(pageParam)
		if err != nil || parsedPage < 1 {
			errors.Add("page", "page must be a positive integer")
		} else {
			page = parsedPage
		}
	}

	limit := defaultLimit
	if strings.TrimSpace(limitParam) != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit < 1 {
			errors.Add("limit", "limit must be a positive integer")
		} else if parsedLimit > maxLimit {
			errors.Add("limit", "limit must be less than or equal to 100")
		} else {
			limit = parsedLimit
		}
	}

	return page, limit, errors
}

func ValidateCreateProject(name string) Errors {
	errors := Errors{}

	if strings.TrimSpace(name) == "" {
		errors.Add("name", "name is required")
	}

	return errors
}

func ValidateUpdateProject(name *string, description *string) Errors {
	errors := Errors{}

	if name == nil && description == nil {
		errors.Add("body", "at least one field must be provided")
		return errors
	}

	if name != nil && strings.TrimSpace(*name) == "" {
		errors.Add("name", "name cannot be empty")
	}

	return errors
}
