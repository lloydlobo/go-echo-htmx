// The services layer coordinates API and database activity to carry out
// application logic.
package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/lloydlobo/go-headcount/internal"
	"github.com/lloydlobo/go-headcount/models"
)

type (
	// Action implements enumeration of actions
	Action int

	// Usage
	//
	// 	var filters = []services.Filter{
	// 			{Url: "#/", Name: "All", Selected: true},
	// 			{Url: "#/active", Name: "Active", Selected: false},
	// 			{Url: "#/completed", Name: "Completed", Selected: false},
	// 	}
	Filter struct {
		Url      string
		Name     string
		Selected bool
	}
)

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

func NewContacts() *ContactService {
	return &ContactService{
		Contacts: models.Contacts{}, // Contact store
		seq:      1,
	}
}

func NewContactsFromAPI() *ContactService {
	var contacts models.Contacts

	apiURL := internal.LookupEnv("API_URL", "https://jsonplaceholder.typicode.com/users")

	contacts, err := fetchUsers(apiURL)
	if err != nil {
		log.Fatalf("failed to fetch and transform users from api: %v", err)
	}

	return &ContactService{
		Contacts: contacts, // Contact store
		seq:      1,
	}
}

type ContactService struct { // Contacts: models.Contacts{}, // Contact store -> *db.ContactStore
	Contacts  models.Contacts // FUTURE: map[int]*Contact
	lock      sync.Mutex      // Lock and defer Unlock during mutation of contacts.
	seq       int             // Tracks times contact is created while server is running. Start from 1.
	idCounter int             // Tracks current count of Contact till when session resets. Start from 0.
}

func (c *ContactService) ResetContacts() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.Contacts = make([]models.Contact, 0)
	c.idCounter = 0
}

func (c *ContactService) CrudOps(action Action, contact models.Contact) models.Contact {
	c.lock.Lock()
	defer c.lock.Unlock()

	index := -1

	// Find index by ID
	if action != ActionCreate {
		for i, r := range c.Contacts {
			if r.ID == contact.ID {
				index = i
				break
			}
		}
	}
	if action == ActionEdit {
		if index == -1 {
			log.Println("error: index is -1", contact)
			return contact
		}
	}

	switch action {
	case ActionCreate:
		c.Contacts = append(c.Contacts, contact)
		c.idCounter += 1
		c.seq += 1
		return contact

	case ActionToggle:
		c.Contacts[index].Status = contact.Status
		return contact

	case ActionUpdate:
		name := strings.Trim(contact.Name, " ")
		phone := strings.Trim(contact.Phone, " ")
		email := strings.Trim(contact.Email, " ") // TODO: add email regexp validation

		if len(name) != 0 && len(phone) != 0 && len(email) != 0 {
			c.Contacts[index].Name = contact.Name
			c.Contacts[index].Email = contact.Email
		} else {
			// remove if name is empty
			c.Contacts = append(c.Contacts[:index], c.Contacts[index+1:]...)
			return models.Contact{}
		}
		return contact

	case ActionDelete:
		c.Contacts = append(c.Contacts[:index], c.Contacts[index+1:]...)

	default:
		// ActionEdit should do nothing but return contact from store
	}

	if index != -1 && action != ActionDelete {
		return c.Contacts[index]
	}

	return models.Contact{} //, errors.Join(errs...)
}

// Get(ctx context.Context, sessionID string)
func (cs *ContactService) Get() (contacts models.Contacts, err error) {
	contacts, err = cs.Contacts, nil

	if len(contacts) == 0 {
		return models.Contacts{}, nil
	}

	return contacts, nil
}

func fetchUsers(apiURL string) (models.Contacts, error) {
	var contacts models.Contacts

	response, err := http.Get(apiURL)
	if err != nil {
		log.Printf("error fetching user data from api: %v", err)
		return contacts, fmt.Errorf("error fetching user data from api: %v", err)
	}
	defer response.Body.Close()

	contactsRaw := []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}{}

	if err := json.NewDecoder(response.Body).Decode(&contactsRaw); err != nil {
		log.Printf("error decoding fetched user data from api: %v", err)
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

	return contacts, nil
}
