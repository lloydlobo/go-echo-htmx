// References:
//
//   - https://github.dev/syarul/todomvc-go-templ-htmx-_hyperscript/blob/main/main.go
//
// Develop:
//
//   - development with hot-module reloading: https://github.com/cosmtrek/air
//     # $ air init
//     # $ air
//   - install templ `go install github.com/a-h/templ/cmd/templ@latest`
//     # $ `templ generate`
//   - create a tailwind.config.js file
//     # $ ./tailwindcss init
//   - start a watcher
//     # $ tailwindcss -i .\templates\css\globals.css -o .\static\css\style.css --watch
//   - compile and minify your CSS for production
//     # $ tailwindcss -i .\templates\css\globals.css -o .\static\css\style.css --minify
//
// Build and Deploy:
//
//   - build command
//     # $ go build -tags netgo -ldflags '-s -w' -o app
//   - pre-deploy command
//     # $ go install github.com/a-h/templ/cmd/templ@latest
//     # $ templ generate
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
	"os"

	"github.com/lloydlobo/go-headcount/handlers"
	"github.com/lloydlobo/go-headcount/internal"
	"github.com/lloydlobo/go-headcount/services"
)

// TODO:
//
//	log := slog.New(slog.NewJSONHandler(os.Stdout))
//	store, err := db.NewContactsStore(os.Getenv("TABLE_NAME"), os.Getenv("AWS_REGION"))
//	cs := services.NewContacts(log, store)
//	h := handlers.New(log, cs)
//	sessionHandler := session.NewMiddleware(h, session.WithSecure(secureFlag))
//	server := &http.Server{Addr: "localhost:8000", Handler: sessionHandler, ReadTimeout: time.Second*10, WriteTimeout: time.Second*10,}
//	server.ListenAndServer()
func runMain() {
	enableGzip := true

	cs := services.NewContacts()
	h := handlers.New(cs)

	// Routes
	http.Handle("/static/", internal.Gzip(http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))))
	http.Handle("/healthcheck", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "200")
	}))

	// Routes::pages
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if enableGzip {
			internal.Gzip(http.HandlerFunc(h.IndexPageHandler)).ServeHTTP(w, r)
		} else {
			http.HandlerFunc(h.IndexPageHandler).ServeHTTP(w, r)
		}
	}))

	// Routes::partials
	http.Handle("/contacts", http.HandlerFunc(h.ContactPartialsHandler))

	// Start the server
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8000"
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Listening on localhost%s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}

func main() {
	runMain()
}
