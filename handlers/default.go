// The handlers layer reads HTTP requests, uses the service to perform CRUD like
// operations, and renders the templ Components.
//
// Errorlog:
//
//   - Note: missing method ServeHTTP
//     cannot use h.IndexPageHandler (value of type func(w http.ResponseWriter, r *http.Request)) as http.Handler value in struct literal: func(w http.ResponseWriter, r *http.Request) does not implement http.Handler (missing method ServeHTTP)compilerInvalidIfaceAssign
//
// Future:
//
//   - Handling non-existent pages:
//
//     What if you visit /view/APageThatDoesntExist? You'll see a page containing HTML. This is because it ignores the error return value from loadPage and continues to try and fill out the template with no data. Instead, if the requested Page doesn't exist, it should redirect the client to the edit Page so the content may be created:
//
//     func viewHandler(w http.ResponseWriter, r *http.Request) {
//     title := r.URL.Path[len("/view/"):]
//     p, err := loadPage(title)
//     if err != nil { http.Redirect(w, r, "/edit/"+title, http.StatusFound); return }
//     renderTemplate(w, "view", p)
//     }
//     The http.Redirect function adds an HTTP status code of http.StatusFound (302) and a Location header to the HTTP response.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/lloydlobo/go-headcount/internal"
	"github.com/lloydlobo/go-headcount/models"
	"github.com/lloydlobo/go-headcount/services"
	"github.com/lloydlobo/go-headcount/templates/components"
	"github.com/lloydlobo/go-headcount/templates/pages"
)

var (
	// Validation Expression to validate title, See Validation, https://go.dev/doc/articles/wiki/
	ValidPath = regexp.MustCompile("^/(about|contacts)/([a-zA-Z0-9]+)$")
)

// ContactService defines the interface for contact-related operations.
type ContactService interface {
	Get() (models.Contacts, error)
	CrudOps(action services.Action, contact models.Contact) models.Contact
	Count() int
	CountByStatus(s models.Status) (count int)
	ResetContacts()
}

// New creates a new DefaultHandler with the given ContactService.
func New(logger *log.Logger, cs ContactService) *DefaultHandler {
	return &DefaultHandler{
		Log:            logger,
		ContactService: cs,
	}
}

// DefaultHandler is a default implementation of the Handler interface.
type DefaultHandler struct {
	Log            *log.Logger
	ContactService ContactService
}

// HandleIndexPage handles requests for GET "/index" page.
func (h *DefaultHandler) HandleIndexPage(w http.ResponseWriter, r *http.Request) {
	if err := h.handleCookieSession(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	indexHTML := pages.IndexPage()
	h.renderView(w, r, indexHTML)
}

// HandleAboutPage handles requests for GET "/about" page.
func (h *DefaultHandler) HandleAboutPage(w http.ResponseWriter, r *http.Request) {
	if err := h.handleCookieSession(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	aboutHTML := pages.AboutPage()
	h.renderView(w, r, aboutHTML)
}

// HandleReadContacts handles requests for contact partials.
//
// HTMX calls this via `<span hx-get="/contacts" hx-target="#hx-contacts" hx-swap="beforeend" hx-trigger="load"></span>`
// So "beforeend" ensures that swap does not mutate the previous elements
func (h *DefaultHandler) HandleReadContacts(w http.ResponseWriter, r *http.Request) {
	contacts, err := h.ContactService.Get()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	h.renderView(w, r, components.ContactsTable(contacts))
}

// HandleReadContact handles HTTP GET - /contacts/{id}.
func (h *DefaultHandler) HandleReadContact(w http.ResponseWriter, r *http.Request) {
	uuidID, err := uuid.Parse(r.PathValue("id")) // Note: Parse should not be used to validate strings as it parses non-standard encodings
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contact := h.ContactService.CrudOps(services.ActionEdit, models.Contact{ID: uuidID})

	w.WriteHeader(http.StatusOK)
	html := components.ContactLi(contact)
	h.renderView(w, r, html)
}

// HandleCreateContact handles HTTP POST - /contacts
func (h *DefaultHandler) HandleCreateContact(w http.ResponseWriter, r *http.Request) {
	contact, err := h.parseContactFromRequestForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.ContactService.CrudOps(services.ActionCreate, contact)

	contacts, err := h.ContactService.Get()
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	html := components.ContactsTable(contacts)
	h.renderView(w, r, html)
}

// HandleGetUpdateContactForm handles HTTP GET - /contacts/{id}/edit.
// Renders a slideout aside with a form pre-filled with contact of id's details.
func (h *DefaultHandler) HandleGetUpdateContactForm(w http.ResponseWriter, r *http.Request) {
	uuidID, err := uuid.Parse(r.PathValue("id")) // Note: Parse should not be used to validate strings as it parses non-standard encodings.
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contact := h.ContactService.CrudOps(services.ActionEdit, models.Contact{ID: uuidID})

	w.WriteHeader(http.StatusOK)
	html := components.Slideout(components.ContactPutForm(contact), "Close", true)
	h.renderView(w, r, html)
}

// HandleUpdateContact handles HTTP PUT - /contacts/{id}.
func (h *DefaultHandler) HandleUpdateContact(w http.ResponseWriter, r *http.Request) {
	contact, err := h.parseContactFromRequestForm(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oldContact := h.ContactService.CrudOps(services.ActionEdit, contact)

	if oldContact.ID != contact.ID {
		err := errors.New("error matching records")
		h.Log.Println(err.Error(), http.StatusNotFound)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	updatedContact := h.ContactService.CrudOps(services.ActionUpdate, contact)

	// If update action returns empty value, incorrect email or contact was deleted,
	// due to empty name (courtesy of todomvc style action.)
	var emptyContact models.Contact
	if updatedContact == emptyContact {
		err := errors.New("something went wrong when updating record")
		h.Log.Println(err.Error(), http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	html := components.ContactLi(updatedContact)
	h.renderView(w, r, html)
}

// HandleDeleteContact handles HTTP DELETE - /contacts/{id}.
//
// To remove the element following a successful DELETE, return a
// 200 status code with an empty body; if the server responds with a 204,
// no swap takes place, documented here: Requests & Responses
//
// Consider options like `hx-swap='none'` for preserving the current state
// or `hx-swap='delete'` for removing elements in response to the request.
func (h *DefaultHandler) HandleDeleteContact(w http.ResponseWriter, r *http.Request) {
	uuidID, err := uuid.Parse(r.PathValue("id")) // Note: Parse should not be used to validate strings as it parses non-standard encodings.
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.ContactService.CrudOps(services.ActionDelete, models.Contact{ID: uuidID})

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "")
}

// HandleGetContactsCount handles HTTP GET requests to /contacts/count
// with optional filtering by active/inactive status.
//
// Filters:
//   - "GET /contacts/count?active=true"
//   - "GET /contacts/count?inactive=true"
func (h *DefaultHandler) HandleGetContactsCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	active, inactive := query.Get(models.StatusActiveQueryKey), query.Get(models.StatusInactiveQueryKey)

	if active != "" && inactive != "" {
		http.Error(w, "invalid query parameters: use either 'active' or 'inactive'", http.StatusBadRequest)
		return
	}

	var count int

	if active == "true" {
		count = h.ContactService.CountByStatus(models.StatusActive)
	} else if inactive == "true" {
		count = h.ContactService.CountByStatus(models.StatusInactive)
	} else {
		count = h.ContactService.Count()
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", count)
}

// TODO: use with central error handling middleware
func (h *DefaultHandler) HandleNotFound(w http.ResponseWriter, r *http.Request) { // 404
	w.WriteHeader(http.StatusNotFound)
	html := pages.NotFoundPage()
	h.renderView(w, r, html)
}

// TODO: use with central error handling middleware
func (h *DefaultHandler) HandlerInternalServerError(w http.ResponseWriter, r *http.Request) { // 500
	w.WriteHeader(http.StatusInternalServerError)
	html := pages.ServerErrorPage()
	h.renderView(w, r, html)
}

func (h *DefaultHandler) HandleHealthcheck(w http.ResponseWriter, r *http.Request) { // "/healthcheck"
	w.Header().Set("Content-Type", "application/json")

	jsonResponse := map[string]string{"status": "ok"}

	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// renderView renders the provided templ.Component to http.ResponseWriter with
// text/html content type.
func (h *DefaultHandler) renderView(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// parseContactFromRequestForm parses contact data from the request form.
func (h *DefaultHandler) parseContactFromRequestForm(r *http.Request) (models.Contact, error) {

	validationError := func(field string, err error) error {
		return fmt.Errorf("validation error for %s: %v", field, err)
	}

	// Extract form values and sanitize them
	id := strings.TrimSpace(html.EscapeString(r.FormValue("id")))
	name := strings.TrimSpace(html.EscapeString(r.FormValue("name")))
	email := strings.TrimSpace(html.EscapeString(r.FormValue("email")))
	phone := strings.TrimSpace(html.EscapeString(r.FormValue("phone")))
	statusRaw := strings.TrimSpace(html.EscapeString(r.FormValue("status")))

	var (
		err    error
		uuidID uuid.UUID
		status models.Status
	)

	// If ID is empty, create a new UUID for POST request; else parse provided ID
	if id == "" {
		uuidID = uuid.New()
	} else {
		if uuidID, err = uuid.Parse(id); err != nil {
			return models.Contact{}, validationError("ID", err)
		}
	}

	if err := internal.ValidateEmail(email); err != nil {
		return models.Contact{}, validationError("Email", err)
	}

	if status, err = (models.StatusParser{}.FormCheckboxValue(statusRaw)); err != nil {
		return models.Contact{}, validationError("Status", err)
	}

	contact := models.Contact{
		ID:     uuidID,
		Name:   name,
		Email:  email,
		Phone:  phone,
		Status: status,
	}

	return contact, nil
}

// handleCookieSession handles session management using cookies.
func (h *DefaultHandler) handleCookieSession(w http.ResponseWriter, r *http.Request) error {
	cookieName := "sessionID"

	_, err := r.Cookie(cookieName)
	if err == http.ErrNoCookie {

		newCookieValue, err := internal.GenRandStr(32)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		}

		newCookie := http.Cookie{
			Name:     cookieName,
			Value:    newCookieValue,
			Expires:  time.Now().Add(time.Second * 6000),
			HttpOnly: true,
		}

		http.SetCookie(w, &newCookie)
		h.ContactService.ResetContacts()
		return nil
	}

	return err
}
