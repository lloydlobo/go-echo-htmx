// The services layer coordinates API and database activity to carry out
// application logic.
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lloydlobo/go-headcount/internal"
	"github.com/lloydlobo/go-headcount/models"
)

// Hack: bypass linter warning for unused function
func TmpInit() {
	cs := NewContactServiceFromAPI()
	cs.updateContactCountCache()
	go func() {
		for {
			time.Sleep(time.Minute)
			cs.updateContactCountCache()
		}
	}()
}

// Action implements enumeration of actions
type Action int

// Enumerate Action related constants in one type
const ( // Hack: using `-1` as `default` case value to act as ActionGet operation.
	ActionCreate Action = iota
	ActionToggle
	ActionEdit
	ActionUpdate
	ActionDelete
)

var (
	ErrUnknownAction error = errors.New("unknown action type")
)

func NewContactService() *ContactService {
	return &ContactService{
		Contacts: models.Contacts{},
		seq:      1,
	}
}

func NewContactServiceFromAPI() *ContactService {
	apiURL := internal.LookupEnv("API_URL", "https://jsonplaceholder.typicode.com/users")

	// Note: Using context.Background() is not idiomatic. Because this is for
	// development and happens before the http server is started.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // cancel context when done fetching

	contacts, err := fetchUsers(ctx, apiURL)
	if err != nil {
		log.Fatalf("failed to fetch and transform users from api: %v", err)
	}

	return &ContactService{
		Contacts: contacts,
		seq:      1,
	}
}

type ContactService struct {
	lock              sync.Mutex // Lock and defer Unlock during mutation of contacts.
	Contacts          models.Contacts
	seq               int // Tracks times contact is created while server is running. Start from 1.
	idCounter         int // Tracks current count of Contact till when session resets. Start from 0.
	ContactCountCache *int64
}

// Get(ctx context.Context, sessionID string)
func (cs *ContactService) Get() (models.Contacts, error) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	if len(cs.Contacts) == 0 {
		return models.Contacts{}, nil
	}

	return cs.Contacts, nil
}

func (cs *ContactService) ResetContacts() {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	cs.Contacts = make([]models.Contact, 0) // OR = nil
	cs.idCounter = 0
}

// FIXME: return (value, error)
func (cs *ContactService) CrudOps(action Action, contact models.Contact) models.Contact {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	index := -1

	if action != ActionCreate {
		index = cs.findIndexByID(contact.ID)

		if index == -1 && action == ActionEdit {
			log.Println("error: index is -1", contact)
			return contact
		}
	}

	switch action {
	case ActionCreate:
		cs.Contacts = append(cs.Contacts, contact)
		cs.idCounter++
		cs.seq++
		// contact.ID = uuid.New() // expect ID to be set by caller
		return contact

	case ActionToggle:
		cs.Contacts[index].Status = contact.Status
		return contact

	case ActionUpdate:
		name := strings.TrimSpace(contact.Name)
		phone := strings.TrimSpace(contact.Phone)
		email := strings.TrimSpace(contact.Email)
		if err := internal.ValidateEmail(email); err != nil {
			log.Printf("failed to validate email: %v", err)
			return models.Contact{} //, fmt.Errorf("error while validating email: %v", err)
		}
		status := contact.Status

		if name != "" && phone != "" && email != "" {
			cs.Contacts[index].Name = name
			cs.Contacts[index].Email = email
			cs.Contacts[index].Phone = phone
			cs.Contacts[index].Status = status
			return contact
		}
		// otherwise remove if name is empty
		cs.deleteContact(index)
		return models.Contact{}

	case ActionDelete:
		cs.deleteContact(index)

	default:
		// ActionEdit should do nothing but return contact from store
	}

	if index != -1 && action != ActionDelete {
		return cs.Contacts[index]
	}

	return models.Contact{} //, errors.Join(errs...)
}

func (cs *ContactService) Count() int {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	return len(cs.Contacts)
}

func (cs *ContactService) CountByStatus(s models.Status) (count int) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	count = 0
	for _, c := range cs.Contacts {
		if c.Status == s {
			count++
		}
	}

	return count
}

func (cs *ContactService) findIndexByID(id uuid.UUID) int {
	for i, c := range cs.Contacts {
		if c.ID == id {
			return i
		}
	}

	return -1
}

func (cs *ContactService) deleteContact(index int) {
	// OR cs.Contacts = append(cs.Contacts[:index], cs.Contacts[index+1:]...)

	if index != -1 {
		_ = copy(cs.Contacts[index:], cs.Contacts[index+1:])
		cs.Contacts = cs.Contacts[:len(cs.Contacts)-1]
	}
}

func fetchUsers(ctx context.Context, apiURL string) (models.Contacts, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	maxRetries := 3
	delay := 1 * time.Second // use exponential backoff for retries instead of fixed count

	var contacts models.Contacts

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			return contacts, fmt.Errorf("error creating request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Attempt %d failed: %v", attempt, err)

			if attempt < maxRetries {
				time.Sleep(delay)
				delay *= 2 // exponential backoff

				continue
			}
			return nil, fmt.Errorf("failed to fetch user data after %d retries: %v", maxRetries, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// transform generic data to models.Contacts
		var contactsRaw []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
			Phone string `json:"phone"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&contactsRaw); err != nil {
			return contacts, fmt.Errorf("error decoding fetched user data from api: %v", err)
		}

		for _, c := range contactsRaw {
			contacts = append(contacts, models.Contact{
				ID:     uuid.New(),
				Name:   c.Name,
				Email:  c.Email,
				Phone:  c.Phone,
				Status: models.StatusInactive,
			})
		}

		return contacts, nil // breaks retries loop
	}

	return nil, errors.New("failed to fetch user data after all retries")
}

// # Usage
//
//	func handleTotalContacts(w http.ResponseWriter, r *http.Request) {
//	  if contactCountCache == nil {
//	    // Handle error or wait for initial cache update
//	    return
//	  }
//	  count := *contactCountCache
//	  // Respond with JSON
//	  json.NewEncoder(w).Encode(map[string]int64{"count": count})
//	}
func (cs *ContactService) updateContactCountCache() {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	n := int64(len(cs.Contacts))
	cs.ContactCountCache = &n
}
