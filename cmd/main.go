package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lloydlobo/go-headcount/handlers"
	"github.com/lloydlobo/go-headcount/internal"
	"github.com/lloydlobo/go-headcount/services"
)

type (
	loggerKeyKind string
)

const (
	buildID                 = "1234567890"
	buildTag                = "v0.0.1"
	loggerKey loggerKeyKind = "logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	logger := log.New(os.Stderr, "", log.LstdFlags)
	ctx = context.WithValue(ctx, loggerKey, logger)

	var cs *services.ContactService
	{
		devENV := true
		if devENV {
			cs = services.NewContactsFromAPI()
		} else {
			cs = services.NewContacts()
		}
	}
	h := handlers.New(logger, cs)
	router := initializeRoutes(logger, h)
	routerWithMiddleware := recoveryMiddleware(router)
	port := internal.LookupEnv("PORT", "1234")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: routerWithMiddleware,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatalf("server error: %v\n", err)
		}
	}()
	logger.Printf("listening on :%s\n", port)

	<-ctx.Done() // Wait for shutdown signal

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("error shutting down server: %v\n", err)
	}
	logger.Println("server gracefully stopped")
}

// initializeRoutes accepts a router instance instead of directly registering routes.
//
// Maybe:
//
//	func initializeRoutes(logger *log.Logger, h *handlers.DefaultHandler) *http.ServeMux {
//	   // ...
//	   // Pages
//	   pages := mux.NewServeMux()
//	   pages.HandleFunc("/", h.IndexPageHandler)
//	   pages.HandleFunc("/about", h.AboutPageHandler)
//	   mux.Handle("/pages/", pages)
//	   // ... (similar groups for partials, etc.)
//	   return mux
//	}
func initializeRoutes(logger *log.Logger, h *handlers.DefaultHandler) *http.ServeMux {
	var withGzip bool = true
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", internal.Gzip(http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))))
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "robots.txt") })

	// Routes for pages
	mux.Handle("/", gzipMiddleware(http.HandlerFunc(h.IndexPageHandler), withGzip))
	mux.Handle("/about", gzipMiddleware(http.HandlerFunc(h.AboutPageHandler), withGzip))

	// Routes for partials
	mux.HandleFunc("POST /contacts", h.HandleCreateContact)
	mux.HandleFunc("GET /contacts", h.HandleReadContacts)
	mux.HandleFunc("GET /contacts/{id}", h.HandleReadContact)
	mux.HandleFunc("PUT /contacts/{id}", h.HandleUpdateContact)
	mux.HandleFunc("DELETE /contacts/{id}", h.HandleDeleteContact)

	mux.HandleFunc("/healthcheck", h.HandleHealthcheck)

	return mux
}

// ----------------------------------------------------------------------------
// MIDDLEWARE

// Fixme: This somehow overides timeout of cancel context
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("application panic: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func gzipMiddleware(next http.Handler, withGzip bool) http.Handler {
	if withGzip {
		return internal.Gzip(next)
	}
	return next
}

// 	Patterns can match the method, host and path of a request. Some examples:
//
// "/index.html" matches the path "/index.html" for any host and method.
// "GET /static/" matches a GET request whose path begins with "/static/".
// "example.com/" matches any request to the host "example.com".
// "example.com/{$}" matches requests with host "example.com" and path "/".
// "/b/{bucket}/o/{objectname...}" matches paths whose first segment is "b" and whose third segment is "o". The name "bucket" denotes the second segment and "objectname" denotes the remainder of the path.
// In general, a pattern looks like
// [METHOD ][HOST]/[PATH]
//
// See Paterns, https://pkg.go.dev/net/http#hdr-Patterns

// Archive
//
// logger := ctx.Value(loggerKey).(*log.Logger)
// fmt.Printf("logger: %v\n", logger)
//
// srv.Handler = mux // Attatch the router to the server
// OR
// handler := RecoveryMiddleware(mux)
// srv.Handler = handler
