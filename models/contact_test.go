package models

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestStatusQueryKeys(t *testing.T) {
	activeKey := strings.ToLower(StatusActive.String())
	inactiveKey := strings.ToLower(StatusInactive.String())

	if StatusActiveQueryKey != activeKey {
		t.Errorf("got %q, want %q", StatusActiveQueryKey, activeKey)
	}

	if StatusInactiveQueryKey != inactiveKey {
		t.Errorf("got %q, want %q", StatusInactiveQueryKey, inactiveKey)
	}
}

func TestStatusIsEnabledAsCheckboxValue(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{"Active", StatusActive, "on"},
		{"Inactive", StatusInactive, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.status.CheckboxValue()
			if err != nil {
				t.Errorf("got %q, want %q, error: %v", got, test.want, err)
			}
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

// All trivial tests below:

func TestContact(t *testing.T) {
	id := uuid.New()
	name := "John Doe"
	email := "john@example.com"
	phone := "1234567890"
	status := StatusActive

	contact := Contact{
		ID:     id,
		Name:   name,
		Email:  email,
		Phone:  phone,
		Status: status,
	}

	if contact.ID != id {
		t.Errorf("got ID %v, want %v", contact.ID, id)
	}

	if contact.Name != name {
		t.Errorf("got Name %s, want %s", contact.Name, name)
	}

	if contact.Email != email {
		t.Errorf("got Email %s, want %s", contact.Email, email)
	}

	if contact.Phone != phone {
		t.Errorf("got Phone %s, want %s", contact.Phone, phone)
	}

	if contact.Status != status {
		t.Errorf("got Status %s, want %s", contact.Status, status)
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{"Active", StatusActive, "Active"},
		{"Inactive", StatusInactive, "Inactive"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.status.String(); got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestStatusIsEnabled(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"Active", StatusActive, true},
		{"Inactive", StatusInactive, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.status.IsEnabled(); got != test.want {
				t.Errorf("got %t, want %t", got, test.want)
			}
		})
	}
}

func TestStatusQueryParam(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{"Active", StatusActive, "active"},
		{"Inactive", StatusInactive, "inactive"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.status.QueryParam(); got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}
