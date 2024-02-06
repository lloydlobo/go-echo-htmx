// The handlers layer reads HTTP requests, uses the service to perform CRUD like
// operations, and renders the templ Components.
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/lloydlobo/go-headcount/internal"
	"github.com/lloydlobo/go-headcount/models"
	"github.com/lloydlobo/go-headcount/services"
	"github.com/lloydlobo/go-headcount/templates/components"
	"github.com/lloydlobo/go-headcount/templates/pages"
)

// ContactService defines the interface for contact-related operations.
type ContactService interface {
	CrudOps(action services.Action, contact models.Contact) models.Contact
	Get() (models.Contacts, error)
	ResetContacts()
}

// New creates a new DefaultHandler with the given ContactService.
//
//	cs := services.NewContacts()
//	h := handlers.New(cs)
func New(cs ContactService) *DefaultHandler { // func New(log *slog.Logger, cs ContactService)
	return &DefaultHandler{ // Log: log,
		ContactService: cs,
	}
}

// DefaultHandler is a default implementation of the Handler interface.
type DefaultHandler struct { // Log *slog.Logger
	ContactService ContactService
}

// IndexPageHandler handles requests for the index page.
func (h *DefaultHandler) IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.handleCookieSession(w, r); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
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

// ContactPartialsHandler handles requests for contact partials.
func (h *DefaultHandler) ContactPartialsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		contacts, err := h.ContactService.Get()

		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// HTMX calls this via <span hx-get="/contacts" hx-target="#hx-contacts" hx-swap="beforeend" hx-trigger="load"></span>
		// So beforeend ensures that swap does not mutate the previous elements
		// PERF: reduce rendering calls and use double buffering like method (collect all <li> and render once at request.)
		for _, contact := range contacts {
			h.RenderView(w, r, components.ContactLi(contact))
		}
		return

	case http.MethodPost:
		contact, err := h.NewContactFromRequestForm(r)
		if err != nil { // Akcshually form value or query error? TODO: use better errors from this method.
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		createdContact := h.ContactService.CrudOps(services.ActionCreate, contact)

		h.RenderView(w, r, components.ContactLi(createdContact))
		return

	default:
		// Note: after implementing all, use -> http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
}

// TODO: seee which h.ContactService.CrudOps(...) can we use for craeting new.
// TODO: move this to services.
func (h *DefaultHandler) newContactFromRequestForm(r *http.Request) (models.Contact, error) {

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	status := r.FormValue("status")

	if err := internal.ValidateEmail(email); err != nil {
		return models.Contact{}, err
	}

	// TODO: escape user input
	contact := models.Contact{
		ID:    uuid.New(),
		Name:  fmt.Sprintf("%v", name),
		Email: fmt.Sprintf("%v", email),
		Phone: fmt.Sprintf("%v", phone),
		Status: func() (s models.Status) {
			if status == "on" {
				return models.StatusActive
			}
			return models.StatusInactive
		}(),
	}

	return contact, nil
}

func (h *DefaultHandler) handleCookieSession(w http.ResponseWriter, r *http.Request) error {
	cookieName := "sessionID"

	_, err := r.Cookie(cookieName)
	if err == http.ErrNoCookie {

		newCookieValue, err := internal.GenRandStr(32)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

// YAGNI: See [HTTP Layer](https://templ.guide/project-structure/project-structure/#http-layer)
type ViewProps struct {
	Filter services.Filter
	// Counts services.Counts
}

// Note that the View method uses the templ Components from the components directory to render the page.
// func (h *DefaultHandler) View(w http.ResponseWriter, r *http.Request, props ViewProps) {
// 		pages.Page(props.Count.Global, props.Counts.Session).Render(r.Context(), w)
// }

// RenderView renders the provided templ.Component to http.ResponseWriter with text/html content type.
func (h *DefaultHandler) RenderView(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}
