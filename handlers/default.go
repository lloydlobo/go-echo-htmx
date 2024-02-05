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

type ContactService interface {
	CrudOps(action services.Action, contact models.Contact) models.Contact
	Seq() int
	ResetContacts()
}

// Usage
//
//	cs := services.NewContacts() // -> &{[] 0 0 {0 0}}
//	h := handlers.New(cs)        // -> &{0xc000022d80}
func New(cs ContactService) *DefaultHandler {
	// func New(log *slog.Logger, cs ContactService)
	return &DefaultHandler{
		// Log: log,
		ContactService: cs,
	}
}

type DefaultHandler struct {
	// Log *slog.Logger
	ContactService ContactService
}

// func (c *ContactsServiceWrapper) pageHandler(w http.ResponseWriter, r *http.Request) {
func (h *DefaultHandler) IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	{ // Note: do we need this now????
		cookieName := "sessionId"

		_, err := r.Cookie(cookieName)

		if err == http.ErrNoCookie { // when err IS ErrNoCookie
			newCookieValue, err := internal.GenRandStr(32)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			newCookie := http.Cookie{
				Name:     cookieName,
				Value:    newCookieValue,
				Expires:  time.Now().Add(time.Second * 6000),
				HttpOnly: true,
			}

			http.SetCookie(w, &newCookie)

			// Start with new contact data when session is reset
			h.ContactService.ResetContacts()
		}
	}

	renderView(w, r, pages.IndexPage())
}

func (h *DefaultHandler) ContactPartialsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		currentContact := h.ContactService.CrudOps(services.Action(-1), models.Contact{})
		renderView(w, r, components.ContactLi(currentContact))
		return
	case http.MethodPost:
		contact := models.Contact{
			ID:    uuid.New(),
			Name:  fmt.Sprintf("John %v", h.ContactService.Seq()) + r.FormValue("name"),
			Email: fmt.Sprintf("john%v@doe.com", h.ContactService.Seq()) + r.FormValue("email"),
			Phone: r.FormValue("phone"),
			Status: func() (status models.Status) {
				if r.FormValue("status") == "on" {
					return models.StatusActive
				}
				return models.StatusInactive
			}(),
		}

		createdContact := h.ContactService.CrudOps(services.ActionCreate, contact)
		renderView(w, r, components.ContactLi(createdContact))
		return
	default:
		// Note: after implementing all, use -> http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
}

// Render the component to http.RespnseWriter and set header content type to text/html.
func renderView(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}
