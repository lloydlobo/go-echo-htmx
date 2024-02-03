package internal

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	// [Reference]:https://pkg.go.dev/github.com/andybalholm/brotli
	// Note: Also see better pkg from same author [matchmaker]: https://pkg.go.dev/github.com/andybalholm/brotli@v1.1.0/matchfinder
	_ "github.com/andybalholm/brotli"
)

// References
//
// https://gist.github.com/bryfry/09a650eb8aac0fb76c24
// https://gist.github.com/CJEnright/bc2d8b8dc0c1389a9feeddb110f822d7
// https://github.com/labstack/echo/blob/master/middleware/compress.go
// https://github.com/nytimes/gziphandler/blob/master/gzip.go
// https://gist.github.com/erikdubbelboer/7df2b2b9f34f9f839a84

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	// wroteHeader bool
	// status      int // status code
}

const (
	gzipScheme = "gzip"

	headerVary            = "Vary"
	headerContentEncoding = "Content-Encoding"
	headerAcceptEncoding  = "Accept-Encoding"
	headerContentType     = "Content-Type"
	headerContentLength   = "Content-Length"
)

/*
{

	if _, ok := w.Header()["Content-Type"]; !ok {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.ResponseWriter.Header().Set("Content-Type", http.DetectContentType(b))
	}
	if !gzr.headerWritten {
		// This is exactly what Go would also do if it hasn't been written yet.
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(b)

}
*/
func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// func (w gzipResponseWriter) WriteHeader(status int) {
// 	w.Header().Del(headerContentLength)
// 	// w.wroteHeader = true
// 	// w.status = status
// 	w.ResponseWriter.WriteHeader(status) // Write status code
// }

// Gzip implements gzip encoding for any http.Handler via "compress/gzip".
//
// Usage
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "text/plain")
//		w.Write([]byte("Hello, World!"))
//	}
//
//	func main() {
//		http.ListenAndServe(":8080", Gzip(handler))
//	}
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get(headerAcceptEncoding), gzipScheme) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set(headerContentEncoding, gzipScheme)

		gw := gzip.NewWriter(w)
		defer gw.Close()

		grw := gzipResponseWriter{ResponseWriter: w, Writer: gw}
		next.ServeHTTP(grw, r)
	})
}

// Usage
//
//	func main() {
//	    mux := http.NewServeMux()
//	    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//	        // Render your HTMX HTML here
//	    })
//	    handler := CompressionMiddleware(mux)
//	    http.ListenAndServe(":8080", handler)
//	}
//
// # References
//
// Header contains the request header fields either received
// by the server or to be sent by the client.
//
// If a server received a request with header lines,
//
//	Host: example.com
//	accept-encoding: gzip, deflate
//	Accept-Language: en-us
//	fOO: Bar
//	foo: two
//
// then
//
//	Header = map[string][]string{
//		"Accept-Encoding": {"gzip, deflate"},
//		"Accept-Language": {"en-us"},
//		"Foo": {"Bar", "two"},
//	}
// func CompressionMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		acceptEncoding := r.Header.Get("Accept-Encoding")

// 		if strings.Contains(acceptEncoding, "br") {
// 			w.Header().Set("Content-Encoding", "br")
// 			w.Header().Set("Vary", "Accept-Encoding") // Indicate content varies based on encoding

// 			// Writes to the returned writer are compressed and written to dst. It is the caller's responsibility to call Close on the Writer when done. Writes may be buffered and not flushed until Close.
// 			bw := brotli.NewWriter(w)
// 			defer bw.Close()

// 			next.ServeHTTP(w, r)
// 		} else if strings.Contains(acceptEncoding, "gzip") {
// 			w.Header().Set("Content-Encoding", "gzip")
// 			w.Header().Set("Vary", "Accept-Encoding") // Indicate content varies based on encoding

// 			// NewWriter returns a new Writer. Writes to the returned writer are compressed and written to w.
// 			gw := gzip.NewWriter(w)
// 			defer gw.Close()

// 			next.ServeHTTP(w, r)
// 		} else {
// 			next.ServeHTTP(w, r)
// 		}
// 	})
// }

/*
package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client supports gzip encoding
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Create a gzip writer
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			// Create a buffer to temporarily store the response
			buffer := &bytes.Buffer{}
			writer := io.MultiWriter(w, buffer)

			// Replace the original ResponseWriter with our custom gzipResponseWriter
			w.Header().Set("Content-Encoding", "gzip")
			w = &gzipResponseWriter{ResponseWriter: w, Writer: io.MultiWriter(gzipWriter, buffer)}

			defer func() {
				// Check if response body is shorter than a certain threshold
				// If so, write the uncompressed response directly to the original ResponseWriter
				if buffer.Len() < 100 { // Replace with your minimum length
					w.Header().Del("Content-Encoding")
					io.Copy(w, buffer)
				}
			}()
		}

		// Continue processing the request
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Implementing the http.ResponseWriter interface
func (w *gzipResponseWriter) WriteHeader(code int) {
	w.Header().Del("Content-Length") // Remove Content-Length header for compressed responses
	w.ResponseWriter.WriteHeader(code)
}

// Implementing the http.ResponseWriter interface
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func main() {
	// Example of using the GzipMiddleware with a simple handler
	http.Handle("/", GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Gzip!"))
	})))

	// Start the server
	http.ListenAndServe(":8080", nil)
}

*/
