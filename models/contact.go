package models

import (
	"strings"

	"github.com/google/uuid"
)

type (
	Contacts []Contact

	Contact struct {
		ID     uuid.UUID `json:"id" form:"id" query:"id"`
		Name   string    `json:"name" form:"name"`
		Email  string    `json:"email" form:"email"`
		Phone  string    `json:"phone" form:"phone"`
		Status Status    `json:"status" form:"status"`
	}

	ContactDTOS struct {
		Name   string
		Email  string
		Phone  string
		Status string // "on" | ""
		// 	   Form status value -> "on" or "" // Can we use json to restrict available strings? like discriminated unions?
	}

	Status string
)

const (
	StatusActive   Status = "Active"
	StatusInactive Status = "Inactive"
)

var (
	StatusActiveQueryKey   = StatusActive.QueryParam()
	StatusInactiveQueryKey = StatusInactive.QueryParam()
)

func (s Status) String() string     { return string(s) }
func (s Status) IsEnabled() bool    { return s == StatusActive }
func (s Status) QueryParam() string { return strings.ToLower(s.String()) }
func (s Status) IsEnabledAsCheckboxValue() string {
	if s == StatusActive {
		return "on"
	}
	return ""
}
