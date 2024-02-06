package internal

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
)

func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}

func GenRandStr(length int) (string, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

const (
	emailRegexExpression = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
)

var (
	// panics if the expression cannot be parsed.
	emailRegexp *regexp.Regexp = regexp.MustCompile(emailRegexExpression)
)

// ValidateEmail checks if the given email address is valid. It combines the `net/mail`
// package's ParseAddress function with a regular expression check.
//
// Note: The `net/mail` package follows RFC 5322 (and extension by RFC 6532) which means it
// may accept email addresses that seem invalid by common standards.
// For instance, local domain names like 't' are accepted. Public domain validity is not verified.
//
// More: https://stackoverflow.com/questions/66624011/how-to-validate-an-email-address-in-go
func ValidateEmail(email string) error {
	// Reference: https://stackoverflow.com/a/77161524
	if len(email) > 320 {
		return fmt.Errorf("email length exceeds 320 characters")
	}

	// Validate using ParseAddress
	emailAddress, err := mail.ParseAddress(email)
	if err != nil || emailAddress.Address != email {
		return err
	}

	// Validate using regular expression
	if !emailRegexp.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil // Email is valid
}
