package internal

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func decodeGzip(input *bytes.Buffer) (string, error) {
	r, err := gzip.NewReader(input)
	if err != nil {
		return "", err
	}
	defer r.Close()

	decoded, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func TestGzipMiddlewareV1(t *testing.T) {
	// Create a simple handler for testing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	})

	// Test case 1: Gzip encoding accepted
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("Accept-Encoding", "gzip")

	rec1 := httptest.NewRecorder()
	Gzip(handler).ServeHTTP(rec1, req1)

	// Ensure response is Gzipped
	if !strings.Contains(rec1.Header().Get("Content-Encoding"), gzipScheme) {
		t.Errorf("Expected Content-Encoding to be gzip, got %s", rec1.Header().Get("Content-Encoding"))
	}

	// Test case 2: Gzip encoding not accepted
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Accept-Encoding", "identity")

	rec2 := httptest.NewRecorder()
	Gzip(handler).ServeHTTP(rec2, req2)

	// Ensure response is not Gzipped
	if rec2.Header().Get("Content-Encoding") != "" {
		t.Errorf("Expected Content-Encoding to be empty, got %s", rec2.Header().Get("Content-Encoding"))
	}

	// Ensure response bodies are the same
	body1, err := decodeGzip(rec1.Body)
	if err != nil {
		t.Errorf("Error decoding Gzip content: %v", err)
	}

	body2 := rec2.Body.String()

	if body1 != body2 {
		t.Errorf("Expected response bodies to be the same, but they differ:\n%s\n%s", body1, body2)
	}
}

func TestGzipMiddlewareV2(t *testing.T) {
	// Create a simple handler for testing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	})

	t.Run("Gzip encoding accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		rec := httptest.NewRecorder()
		Gzip(handler).ServeHTTP(rec, req)

		t.Run("Content-Encoding should be gzip", func(t *testing.T) {
			if got := rec.Header().Get("Content-Encoding"); !strings.Contains(got, gzipScheme) {
				t.Errorf("Expected Content-Encoding to be gzip, got %s", got)
			}
		})

		t.Run("Response body should be decoded and match original", func(t *testing.T) {
			body, err := decodeGzip(rec.Body)
			if err != nil {
				t.Errorf("Error decoding Gzip content: %v", err)
			}

			expectedBody := "Hello, World!"
			if body != expectedBody {
				t.Errorf("Expected response body to be %s, but got %s", expectedBody, body)
			}
		})
	})

	t.Run("Gzip encoding not accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "identity")

		rec := httptest.NewRecorder()
		Gzip(handler).ServeHTTP(rec, req)

		t.Run("Content-Encoding should be empty", func(t *testing.T) {
			if got := rec.Header().Get("Content-Encoding"); got != "" {
				t.Errorf("Expected Content-Encoding to be empty, got %s", got)
			}
		})

		t.Run("Response body should match original", func(t *testing.T) {
			body := rec.Body.String()

			expectedBody := "Hello, World!"
			if body != expectedBody {
				t.Errorf("Expected response body to be %s, but got %s", expectedBody, body)
			}
		})
	})
}
