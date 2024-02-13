package models

import (
	"fmt"
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
	}
)

type Status string

const (
	StatusActive   Status = "Active"
	StatusInactive Status = "Inactive"
	StatusError    Status = "Error" // Sentinel value for unexpected status
)

var (
	StatusActiveQueryKey   = StatusActive.QueryParam()
	StatusInactiveQueryKey = StatusInactive.QueryParam()
)

func (s Status) String() string     { return string(s) }
func (s Status) IsEnabled() bool    { return s == StatusActive }
func (s Status) QueryParam() string { return strings.ToLower(s.String()) }

func (s Status) CheckboxValue() (string, error) {
	switch s {
	case StatusActive:
		return "on", nil
	case StatusInactive:
		return "", nil
	default:
		return string(StatusError), fmt.Errorf("unexpected status: %v", s)
	}
}

type StatusParser struct {
	Status Status
}

func (sh StatusParser) FormCheckboxValue(s string) (status Status, err error) {
	switch strings.TrimSpace(s) {
	case "on":
		return StatusActive, nil
	case "":
		return StatusInactive, nil
	default:
		return StatusError, fmt.Errorf("invalid checkbox value for status: %s", s)
	}
}
