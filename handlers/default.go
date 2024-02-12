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
	CrudOps(action services.Action, contact models.Contact) models.Contact
	Get() (models.Contacts, error)
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

// IndexPageHandler handles requests for GET "/index" page.
func (h *DefaultHandler) IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.handleCookieSession(w, r); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	indexHTML := pages.IndexPage()
	h.RenderView(w, r, indexHTML)
}

// AboutPageHandler handles requests for GET "/about" page.
func (h *DefaultHandler) AboutPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.handleCookieSession(w, r); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	aboutHTML := pages.AboutPage()
	h.RenderView(w, r, aboutHTML)
}

// HandleReadContacts handles requests for contact partials.
//
// HTMX calls this via `<span hx-get="/contacts" hx-target="#hx-contacts" hx-swap="beforeend" hx-trigger="load"></span>`
// So "beforeend" ensures that swap does not mutate the previous elements
func (h *DefaultHandler) HandleReadContacts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.Log.Printf("expected %s but got %s", http.MethodGet, r.Method)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	contacts, err := h.ContactService.Get()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// PERF: reduce rendering calls and use double buffering like method (collect all <li> and render once at request.)

	// for _, contact := range contacts {
	// 	h.RenderView(w, r, components.ContactLi(contact))
	// }
	h.RenderView(w, r, components.ContactsTable(contacts))
}

// HandleReadContact handles HTTP GET - /contacts/{id}.
func (h *DefaultHandler) HandleReadContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.Log.Fatalf("expected %s but got %s", http.MethodGet, r.Method)
	}

	// Note: Parse should not be used to validate strings as it parses non-standard encodings
	uuidID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		// 	return ctx.JSON(http.StatusNotAcceptable, map[string]string{"message": "Unparsable parameter"})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contact := h.ContactService.CrudOps(services.ActionEdit, models.Contact{ID: uuidID})

	w.WriteHeader(http.StatusOK)
	html := components.ContactLi(contact)
	h.RenderView(w, r, html)
}

// HandleCreateContact handles HTTP POST - /contacts
func (h *DefaultHandler) HandleCreateContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.Log.Fatalf("expected %s but got %s", http.MethodPost, r.Method)
	}

	contact, err := h.parseContactFromRequestForm(r)
	if err != nil { // Akcshually form value or query error? TODO: use better errors from this method.
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
	h.RenderView(w, r, html)
}

// HandleGetUpdateContactForm handles HTTP GET - /contacts/{id}/edit.
// Renders a slideout aside with a form pre-filled with contact of id's details.
func (h *DefaultHandler) HandleGetUpdateContactForm(w http.ResponseWriter, r *http.Request) {
	rawid := r.PathValue("id")

	id, err := uuid.Parse(rawid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contact := h.ContactService.CrudOps(services.ActionEdit, models.Contact{ID: id})

	w.WriteHeader(http.StatusOK)
	html := components.Slideout(components.ContactPutForm(contact), "Close", true)
	h.RenderView(w, r, html)
}

// HandleUpdateContact handles HTTP PUT - /contacts/{id}.
//
// Reference: https://htmx.org/examples/update-other-content/#events
//
//	<tbody id="contacts-table" hx-get="/contacts/table" hx-trigger="newContact from:body">
//
// When a successful contact creation occurs during a POST to /contacts, the
// response includes an HX-Trigger response header that looks like this:
//
// HX-Trigger:newContact
// This will trigger the table to issue a GET to /contacts/table and this
// will render the newly added contact row (in addition to the rest of the table.)
func (h *DefaultHandler) HandleUpdateContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.Log.Fatalf("expected %s but got %s", http.MethodPut, r.Method)
	}

	// FIXME: error while validating email: mail: no address
	contact, err := h.parseContactFromRequestForm(r)
	if err != nil {
		h.Log.Println("failed to parse contact from request form", err.Error(), http.StatusInternalServerError)
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

	var emptyContact models.Contact
	if updatedContact == emptyContact {
		err := errors.New("something went wrong when updating record") // if update action returns empty value, incorrect email or contact was deleted due to empty name(courtesy of todomvc style action.)
		h.Log.Println(err.Error(), http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// FIXME: implement this to update the whole table on newContact event.
	//     this ensures that data stays uptodate and doesnt remain stale(multi users sessions.)
	w.Header().Add("HX-Trigger", "newContact")
	w.WriteHeader(http.StatusOK)
	html := components.ContactLi(updatedContact)
	h.RenderView(w, r, html)
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
	if r.Method != http.MethodDelete {
		h.Log.Fatalf("expected %s but got %s", http.MethodDelete, r.Method)
	}

	// Note: Parse should not be used to validate strings as it parses non-standard encodings
	uuidID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.ContactService.CrudOps(services.ActionDelete, models.Contact{ID: uuidID})

	// TODO!!!!!!!!!!!!!!!!!!!!!!!!!1: put the table of contacts inside a form
	// remove hx-target or swap with body:. see htmx docs
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{}) // or use fmt.Fprintf(w, "")
}

// --------------------------------------------------------------------------------------------------

// TODO: use with central error handling middleware
func (h *DefaultHandler) HandleNotFound(w http.ResponseWriter, r *http.Request) { // 404
	w.WriteHeader(http.StatusNotFound)
	html := pages.NotFoundPage()
	h.RenderView(w, r, html)
}

// TODO: use with central error handling middleware
func (h *DefaultHandler) HandlerInternalServerError(w http.ResponseWriter, r *http.Request) { // 500
	w.WriteHeader(http.StatusInternalServerError)
	html := pages.ServerErrorPage()
	h.RenderView(w, r, html)
}

func (h *DefaultHandler) HandleHealthcheck(w http.ResponseWriter, r *http.Request) { // "/healthcheck"
	w.Header().Set("Content-Type", "application/json")

	jsonResponse := map[string]string{"status": "ok"}

	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// YAGNI: See [HTTP Layer](https://templ.guide/project-structure/project-structure/#http-layer)
type ViewProps struct {
	Filter services.Filter
	// Counts services.Counts
}

// Note that the View method uses the templ Components from the components directory to render the page.
// func (h *DefaultHandler) View(w http.ResponseWriter, r *http.Request, props ViewProps) {
// 		pages.Page(props.Count.Global, props.Counts.Session).Render(r.Context(), w)
// }

// RenderView renders the provided templ.Component to http.ResponseWriter with
// text/html content type.
//
// Handle the errors and return an error message to the user. That way if
// something does go wrong, the server will function exactly how we want and
// the user can be notified.
func (h *DefaultHandler) RenderView(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *DefaultHandler) parseContactFromRequestForm(r *http.Request) (models.Contact, error) {
	id := strings.TrimSpace(html.EscapeString(r.FormValue("id")))
	name := strings.TrimSpace(html.EscapeString(r.FormValue("name")))
	email := strings.TrimSpace(html.EscapeString(r.FormValue("email")))
	phone := strings.TrimSpace(html.EscapeString(r.FormValue("phone")))
	statusRaw := strings.TrimSpace(html.EscapeString(r.FormValue("status")))

	var uuidID uuid.UUID
	if id == "" {
		uuidID = uuid.New() // for a POST request
	} else {
		var err error
		uuidID, err = uuid.Parse(id)
		if err != nil {
			return models.Contact{}, fmt.Errorf("error parsing id: %v", err)
		}
	}

	if err := internal.ValidateEmail(email); err != nil {
		return models.Contact{}, fmt.Errorf("error parsing email: %v", err)
	}

	var status models.Status
	switch statusRaw {
	case "on":
		status = models.StatusActive
	case "":
		status = models.StatusInactive
	default:
		return models.Contact{}, fmt.Errorf("unexpected form value for 'status': %s", statusRaw)
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
