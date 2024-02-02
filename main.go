// References:
//
//   - https://github.dev/syarul/todomvc-go-templ-htmx-_hyperscript/blob/main/main.go
//   - https://github.com/cosmtrek/air
//     development:
//     # $ air init
//     # $ air
//   - install templ `go install github.com/a-h/templ/cmd/templ@latest`
//     # $ `templ generate`
package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

// TODO: Export toutes as json!!

type (
	Filter struct {
		url      string
		name     string
		selected bool
	}

	// Action implements enumeration of actions
	Action int
)

// Enumerate Action related constants in one type
const (
	Create Action = iota
	Toggle
	Edit
	Update
	Delete
)

var (
	filters = []Filter{
		{url: "#/", name: "All", selected: true},
		{url: "#/active", name: "Active", selected: false},
		{url: "#/completed", name: "Completed", selected: false},
	}

	// Tracks current count of Contact
	idCounter uint64
)

func main() {
	runMain()
}

func templRenderer(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component.Render(r.Context(), w)
}

func (c *Contacts) crudOps(action Action, contact Contact) Contact {
	index := -1

	if action != Create {
		for i, r := range *c {
			if r.ID == contact.ID {
				index = i
				break
			}
		}
	}

	switch action {
	case Create:
		*c = append(*c, contact)
		return contact
	case Toggle:
		(*c)[index].Status = contact.Status
		return contact
	case Update:
		name := strings.Trim(contact.Name, " ")
		phone := strings.Trim(contact.Phone, " ")
		// TODO: add email regexp validation
		email := strings.Trim(contact.Email, " ")

		if len(name) != 0 && len(phone) != 0 && len(email) != 0 {
			(*c)[index].Name = contact.Name
			(*c)[index].Email = contact.Email
		} else {
			// Remove if name is empty
			*c = append((*c)[:index], (*c)[index+1:]...)
			return Contact{}
		}
		return contact
	case Delete:
		*c = append((*c)[:index], (*c)[index+1:]...)
	default:
		// Edit should do nothing but retuurn contact from store
	}

	if index != (-1) && action != Delete {
		return (*c)[index]
	}

	return Contact{}
}

func runMain() {
	flagUnimplemented := false

	c := InitializeModels()
	c = append(c, Contact{ID: uuid.New(), Name: "John Doe", Email: "john@doe.com", Phone: "1234567890", Status: StatusActive})

	// SetupRoutes()

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "200")
	}

	http.Handle("/healthcheck", http.HandlerFunc(handler))
	http.Handle("/", http.HandlerFunc(c.pageHandler))

	// Serve *._hs (Hyperscript) files
	http.Handle("/hs/", http.StripPrefix("/hs/", http.FileServer(http.Dir("./hs"))))

	// use the http.Handle to register the file server handler for a specific route
	if flagUnimplemented { // this is used to serve axe-core for the todomvc test. [See](https://github.dev/syarul/todomvc-go-templ-htmx-_hyperscript/blob/main/main.go)
		dir := "./cypress-example=todomvc/node_modules"
		http.Handle("/node_modules/", http.StripPrefix("/node_modules/", http.FileServer(http.Dir(dir))))
	}

	// start the server
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8000"
	}

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("Server is running on http://localhost%s\n", addr)

	// Start the HTTP server
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

func InitializeModels() Contacts {
	contacts := []Contact{}
	return contacts
}

//---------
// HANDLERS
//---------

// Naive handler:
//
//	func (c *Contacts) pageHandler(w http.ResponseWriter, r *http.Request) {
//		x := fmt.Sprintf("Hello, this is a simple Go server!\n%+v", c)
//		fmt.Fprintln(w, x)
//	}
func (c *Contacts) pageHandler(w http.ResponseWriter, r *http.Request) {
	cookieName := "seesionId"

	_, err := r.Cookie(cookieName)

	if err == http.ErrNoCookie { // when err IS ErrNoCookie
		newCookieValue, err := genRandStr(32)

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

		// Start with new contact data when session is resest
		*c = make([]Contact, 0)
		// idContact=0
	}

	// TODO:
	// 	templRenderer(w, r, Page(*t, filters, defChecked(*t), hasCompleteTask(*t), selectedFilter(filters)))
	templRenderer(w, r, Page(*c, filters))

	// tmpl, err := template.ParseFiles("index.html")
	// if err != nil {
	// 	fmt.Println("Error parsing template:", err)
	// 	return
	// }
	//
	// data := struct{ Name string }{"John"}
	// if err := tmpl.Execute(w, data); err != nil {
	// 	fmt.Println("Error executing template:", err)
	// }
}

//---------------
// RAND GEN UTILS
//---------------

// TODO: test me
func genRandStr(length int) (string, error) {
	bytes := make([]byte, length)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
