package services

import (
	"strings"
	"sync"

	"github.com/lloydlobo/go-headcount/models"
)

type (
	Filter struct {
		Url      string
		Name     string
		Selected bool
	}

	// Action implements enumeration of actions
	Action int
)

// Enumerate Action related constants in one type
const ( // Hack: using `-1` as `default` case value to act as ActionGet operation.
	ActionCreate Action = iota
	ActionToggle
	ActionEdit
	ActionUpdate
	ActionDelete
)

type ContactService struct {
	Contacts  models.Contacts // map[int]*Contact
	seq       int
	idCounter int
	lock      sync.Mutex
}

func NewContacts() *ContactService {
	return &ContactService{
		Contacts: models.Contacts{},
	}
}

func (c *ContactService) Seq() int {
	return c.seq
}

func (c *ContactService) ResetContacts() {
	c.Contacts = make([]models.Contact, 0)
}

func (c *ContactService) ResetIdCounter() {
	c.idCounter = 0
}

func (c *ContactService) CrudOps(action Action, contact models.Contact) models.Contact {
	index := -1

	if action != ActionCreate {
		for i, r := range c.Contacts {
			if r.ID == contact.ID {
				index = i
				break
			}
		}
	}

	switch action {
	case ActionCreate:
		c.lock.Lock()
		defer c.lock.Unlock()

		c.Contacts = append(c.Contacts, contact)
		c.idCounter += 1
		c.seq += 1

		return contact
	case ActionToggle:
		c.lock.Lock()
		defer c.lock.Unlock()

		(c.Contacts)[index].Status = contact.Status

		return contact
	case ActionUpdate:
		c.lock.Lock()
		defer c.lock.Unlock()

		name := strings.Trim(contact.Name, " ")
		phone := strings.Trim(contact.Phone, " ")
		email := strings.Trim(contact.Email, " ") // TODO: add email regexp validation

		if len(name) != 0 && len(phone) != 0 && len(email) != 0 {
			(c.Contacts)[index].Name = contact.Name
			(c.Contacts)[index].Email = contact.Email
		} else { // remove if name is empty
			c.Contacts = append((c.Contacts)[:index], (c.Contacts)[index+1:]...)
			return models.Contact{}
		}

		return contact
	case ActionDelete:
		c.lock.Lock()
		defer c.lock.Unlock()

		c.Contacts = append((c.Contacts)[:index], (c.Contacts)[index+1:]...)
	default:
		// edit should do nothing but return contact from store
	}

	if index != (-1) && action != ActionDelete {
		c.lock.Lock()
		defer c.lock.Unlock()

		return (c.Contacts)[index]
	}

	return models.Contact{}
}
