// Develop:
//
//   - install dependencies
//     # $ go install github.com/a-h/templ/cmd/templ@latest
//
//   - development with hot-module reloading: https://github.com/cosmtrek/air
//     # $ air init
//     # $ air
//
//   - install templ `go install github.com/a-h/templ/cmd/templ@latest`
//     # $ `templ generate`
//
//   - create a tailwind.config.js file
//     # $ ./tailwindcss init
//
//   - start a watcher
//     # $ tailwindcss -i .\templates\css\globals.css -o .\static\css\style.css --watch
//
//   - compile and minify your CSS for production
//     # $ tailwindcss -i .\templates\css\globals.css -o .\static\css\style.css --minify
//
// Build and Deploy:
//
//   - pre-build commands
//     # $ tailwindcss -i .\templates\css\globals.css -o .\static\css\style.css --minify
//     # $ templ generate
//
//   - build command
//     # $ go build -tags netgo -ldflags '-s -w' -o app
//
// References:
//
//   - https://github.dev/syarul/todomvc-go-templ-htmx-_hyperscript/blob/main/main.go
//
// Errorlog:
//
//   - Error: listen tcp :8000: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.
//     >>> While spamming POST "/contacts" -> should rate limit
//     >>> Seems like `air` dev tool error.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lloydlobo/go-headcount/handlers"
	"github.com/lloydlobo/go-headcount/internal"
	"github.com/lloydlobo/go-headcount/services"
)

func runMain() {
	enableGzip := true

	cs := services.NewContacts()
	h := handlers.New(cs)

	// Serve static files
	http.Handle("/static/", internal.Gzip(http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))))
	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "robots.txt") })

	// Routes
	http.Handle("/healthcheck", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, http.StatusOK) }))

	// Routes for pages
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if enableGzip {
			internal.Gzip(http.HandlerFunc(h.IndexPageHandler)).ServeHTTP(w, r)
		} else {
			http.HandlerFunc(h.IndexPageHandler).ServeHTTP(w, r)
		}
	}))

	// Routes for partials
	http.Handle("/contacts", http.HandlerFunc(h.ContactPartialsHandler))

	port := internal.LookupEnv("PORT", "8000")
	log.Printf("Listening on localhost:%s\n", port)

	// Start the server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}

func main() {
	runMain()
}
