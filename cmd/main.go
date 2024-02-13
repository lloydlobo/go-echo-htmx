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
	port := internal.LookupEnv("PORT", "1234")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	logger := log.New(os.Stderr, "HTTP ", log.LstdFlags)
	ctx = context.WithValue(ctx, loggerKey, logger)

	cs := services.NewContactServiceFromAPI()
	h := handlers.New(logger, cs)
	router := initializeRoutes(h)
	routerWithMiddleware := recoveryMiddleware(router)

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
// Patterns can match the method, host and path of a request. See Paterns, https://pkg.go.dev/net/http#hdr-Patterns
// [METHOD ][HOST]/[PATH]
func initializeRoutes(h *handlers.DefaultHandler) *http.ServeMux {
	mux := http.NewServeMux()
	
	// Serve static files
	mux.Handle("/static/", internal.Gzip(http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))))
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "robots.txt") })
	
	// Routes for pages
	var withGzip bool = true // flag
	mux.Handle("/", gzipMiddleware(http.HandlerFunc(h.HandleIndexPage), withGzip))
	mux.Handle("/about", gzipMiddleware(http.HandlerFunc(h.HandleAboutPage), withGzip))

	// Routes for partials
	mux.HandleFunc("POST /contacts", h.HandleCreateContact)
	mux.HandleFunc("GET /contacts", h.HandleReadContacts)
	mux.HandleFunc("GET /contacts/{id}", h.HandleReadContact)
	mux.HandleFunc("PUT /contacts/{id}", h.HandleUpdateContact)
	mux.HandleFunc("DELETE /contacts/{id}", h.HandleDeleteContact)
	mux.HandleFunc("GET /contacts/count", h.HandleGetContactsCount)
	mux.HandleFunc("GET /contacts/count?active=true", h.HandleGetContactsCount)
	mux.HandleFunc("GET /contacts/count?inactive=true", h.HandleGetContactsCount)

	// Routes for intermediate requests
	mux.HandleFunc("GET /contacts/{id}/edit", h.HandleGetUpdateContactForm)

	mux.HandleFunc("/healthcheck", h.HandleHealthcheck)

	return mux
}

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
