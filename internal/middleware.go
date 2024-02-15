// References
//
// https://gist.github.com/bryfry/09a650eb8aac0fb76c24
// https://gist.github.com/CJEnright/bc2d8b8dc0c1389a9feeddb110f822d7
// https://github.com/labstack/echo/blob/master/middleware/compress.go
// https://github.com/nytimes/gziphandler/blob/master/gzip.go
// https://gist.github.com/erikdubbelboer/7df2b2b9f34f9f839a84

package internal

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

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

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Note: unimplemented
func (w gzipResponseWriter) WriteHeader(status int) {
	w.Header().Del(headerContentLength)
	// w.wroteHeader = true
	// w.status = status
	w.ResponseWriter.WriteHeader(status) // Write status code
}
