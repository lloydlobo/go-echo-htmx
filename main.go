// References:
//
//   - https://github.dev/syarul/todomvc-go-templ-htmx-_hyperscript/blob/main/main.go
//   - https://github.com/cosmtrek/air
//     development:
//     # $ air init
//     # $ air
//   - install templ `go install github.com/a-h/templ/cmd/templ@latest`
//     # $ `templ generate`
//   - build command
//     # $ go build -tags netgo -ldflags '-s -w' -o app
//   - pre-deploy command
//     # $ go install github.com/a-h/templ/cmd/templ@latest
//     # $ templ generate
//   - Create a tailwind.config.js file
//     # $ ./tailwindcss init
//   - Start a watcher
//     # $ tailwindcss -i globals.css -o .\static\css\style.css --watch
//   - Compile and minify your CSS for production
//     # $ tailwindcss -i globals.css -o .\static\css\style.css --minify
//
// Errorlog:
//
//   - Error: listen tcp :8000: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.
//     >>> While spamming POST "/contacts" -> should rate limit
//
// TODO: Export routes as json!!
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"

	"github.com/lloydlobo/go-headcount/components"
	"github.com/lloydlobo/go-headcount/internal"
	"github.com/lloydlobo/go-headcount/model"
	"github.com/lloydlobo/go-headcount/services"
)

var (
	filters = []services.Filter{
		{Url: "#/", Name: "All", Selected: true},
		{Url: "#/active", Name: "Active", Selected: false},
		{Url: "#/completed", Name: "Completed", Selected: false},
	}

	idCounter uint64         // Tracks current count of Contact till when session resets. Start from 0.
	seq       = 1            // Tracks times contact is created while server is running. Start from 1.
	lock      = sync.Mutex{} // Lock and defer Unlock during mutation of contacts
)

// Render the component to http.RespnseWriter and set header content type to text/html.
func RenderView(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}

//---------
// SERVICES
//---------

type (
	ContactsServiceWrapper struct {
		Contacts *model.Contacts
	}
)

// TODO: Move this to services/
func (c *ContactsServiceWrapper) CrudOps(action services.Action, contact model.Contact) model.Contact {
	index := -1

	if action != services.ActionCreate {
		for i, r := range *c.Contacts {
			if r.ID == contact.ID {
				index = i
				break
			}
		}
	}

	switch action {
	case services.ActionCreate:
		lock.Lock()
		defer lock.Unlock()
		*c.Contacts = append(*c.Contacts, contact)
		idCounter += 1
		seq += 1
		return contact
	case services.ActionToggle:
		lock.Lock()
		defer lock.Unlock()
		(*c.Contacts)[index].Status = contact.Status
		return contact
	case services.ActionUpdate:
		lock.Lock()
		defer lock.Unlock()
		name := strings.Trim(contact.Name, " ")
		phone := strings.Trim(contact.Phone, " ")
		email := strings.Trim(contact.Email, " ") // TODO: add email regexp validation
		if len(name) != 0 && len(phone) != 0 && len(email) != 0 {
			(*c.Contacts)[index].Name = contact.Name
			(*c.Contacts)[index].Email = contact.Email
		} else {
			// remove if name is empty
			*c.Contacts = append((*c.Contacts)[:index], (*c.Contacts)[index+1:]...)
			return model.Contact{}
		}
		return contact
	case services.ActionDelete:
		lock.Lock()
		defer lock.Unlock()
		*c.Contacts = append((*c.Contacts)[:index], (*c.Contacts)[index+1:]...)
	default:
		// edit should do nothing but return contact from store
	}

	if index != (-1) && action != services.ActionDelete {
		lock.Lock()
		defer lock.Unlock()
		return (*c.Contacts)[index]
	}

	return model.Contact{}
}

//---------
// HANDLERS
//---------

// contactsHandler handles GET|POST request to "/contacts".
func (c *ContactsServiceWrapper) contactsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		lock.Lock()
		defer lock.Unlock()
		if len(*c.Contacts) == 0 {
			fmt.Fprintln(w, nil) // templateString.Execute(w, nil)
			return
		}
		currentContact := (*c.Contacts)[0]
		RenderView(w, r, components.ContactLi(currentContact))
		return
	case http.MethodPost:
		contact := model.Contact{
			ID:    uuid.New(),
			Name:  fmt.Sprintf("John %v", seq) + r.FormValue("name"),
			Email: fmt.Sprintf("john%v@doe.com", seq) + r.FormValue("email"),
			Phone: r.FormValue("phone"),
			Status: func() (status model.Status) {
				if r.FormValue("status") == "on" {
					return model.StatusActive
				}
				return model.StatusInactive
			}(),
		} // log.Println(seq, idCounter, contact)
		createdContact := c.CrudOps(services.ActionCreate, contact)
		RenderView(w, r, components.ContactLi(createdContact))
		return
	default: // http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		return
	}
}

// pageHandler implements rendering `templ` index.html for `GET` at route `"/"`.
//
// Naive handler:
//
//	func (c *Contacts) pageHandler(w http.ResponseWriter, r *http.Request) {
//		x := fmt.Sprintf("Hello, this is a simple Go server!\n%+v", c)
//		fmt.Fprintln(w, x)
//	}
//
// Naive handler with template:
//
//	tmpl, err := template.ParseFiles("index.html")
//	if err != nil {
//		fmt.Println("Error parsing template:", err)
//		return
//	}
//	data := struct{ Name string }{"John"}
//	if err := tmpl.Execute(w, data); err != nil {
//		fmt.Println("Error executing template:", err)
//	}
func (c *ContactsServiceWrapper) pageHandler(w http.ResponseWriter, r *http.Request) {
	cookieName := "seesionId"
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
		*c.Contacts = make([]model.Contact, 0)
		idCounter = 0
	}

	// TODO: 	templRenderer(w, r, Page(*t, filters, defChecked(*t), hasCompleteTask(*t), selectedFilter(filters)))
	RenderView(w, r, components.Page(*c.Contacts, filters))
}

func initializeModels() model.Contacts {
	contacts := []model.Contact{}
	return contacts
}

func runMain() {
	flagWithGzip := true // TODO: move to Config

	initialContacts := initializeModels()
	// log := slog.New(slog.NewJSONHandler(os.Stdout))
	// store, err := db.NewContactsStore(os.Getenv("TABLE_NAME"), os.Getenv("AWS_REGION"))
	// cs := services.NewContacts(log, store)
	// h := handlers.New(log, cs)
	c := ContactsServiceWrapper{Contacts: &initialContacts}
	// sessionHandler := session.NewMiddleware(h, session.WithSecure(secureFlag))
	// server := &http.Server{Addr: "localhost:8000", Handler: sessionHandler, ReadTimeout: time.Second*10, WriteTimeout: time.Second*10,}
	// server.ListenAndServer()

	// Routes: SetupRoutes()
	handler := func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "200") }
	http.Handle("/healthcheck", http.HandlerFunc(handler))
	if flagWithGzip { // with gzip 1.2 kB | min=3m startup_max=63ms
		http.Handle("/", internal.Gzip(http.HandlerFunc(c.pageHandler)))
	} else { // without gzip 2.6 kB | min=3ms startup_max=46ms
		http.Handle("/", http.HandlerFunc(c.pageHandler))
	}
	http.Handle("/contacts", http.HandlerFunc(c.contactsHandler))

	// serve static folder assets
	http.Handle("/static/", internal.Gzip(http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))))

	// start the server
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8000"
	}
	addr := fmt.Sprintf(":%s", port)

	// Start the HTTP server
	fmt.Printf("Listening on localhost%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

func main() {
	runMain()
}
