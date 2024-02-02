package main

import "github.com/google/uuid"

type (
	Contact struct {
		ID     uuid.UUID `json:"id" form:"id" query:"id"`
		Name   string    `json:"name" form:"name"`
		Email  string    `json:"email" form:"email"`
		Phone  string    `json:"phone" form:"phone"`
		Status Status    `json:"status" form:"status"`
	}

	Status string

	Contacts []Contact
)

const (
	StatusActive   Status = "Active"
	StatusInactive Status = "Inactive"
)

func (s Status) String() string  { return string(s) }
func (s Status) IsEnabled() bool { return s == StatusActive }
