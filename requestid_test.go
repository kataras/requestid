package requestid

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWithGenerator(t *testing.T) {
	const expected = "my_id"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(xRequestIDHeaderKey, expected)
	defer r.Body.Close()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Get(r)
		fmt.Fprint(w, id)
	})

	gen := func(w http.ResponseWriter, r *http.Request) string {
		id := r.Header.Get(xRequestIDHeaderKey)
		w.Header().Set(xRequestIDHeaderKey, id)
		return id
	}

	HandlerWithGenerator(h, gen)(w, r)
	if got := w.Body.String(); expected != got {
		t.Fatalf("expected id: '%s' but got: '%s'", expected, got)
	}

	// test if response header set.
	if got := w.Header().Get(xRequestIDHeaderKey); expected != got {
		t.Fatalf("expected response header: '%s' but got: '%s'", expected, got)
	}
}

func TestHandlerCustomID(t *testing.T) {
	const expected = "custom_id"
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	defer r.Body.Close()

	setCustomMiddleware := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r = Set(r, expected)
			next.ServeHTTP(w, r)
		}
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Get(r)
		fmt.Fprint(w, id)
	})

	Handler(setCustomMiddleware(h))(w, r)
	if got := w.Body.String(); expected != got {
		t.Fatalf("expected id: '%s' but got: '%s'", expected, got)
	}
}
