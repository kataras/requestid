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

func TestHandlerWithHashGenerator(t *testing.T) {
	expected := "ba5ce1ba53e508a854ca039833eb33c872e895cd"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	defer r.Body.Close()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := Get(r)
		fmt.Fprint(w, id)
	})

	HandlerWithGenerator(h, HashGenerator(false))(w, r)
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

func TestHash(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Body.Close()

	r.Header.Set("User-Agent", "Mozilla/5.0 (Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0")
	hash1 := Hash(r, false)
	if expected, got := "a128650e109e98ff41c684cc3543f11130fa9c6b", hash1; expected != got {
		t.Fatalf("[1] expected request's hash to be: '%s' but got: '%s'", expected, got)
	}

	r.Header.Set("User-Agent", "mozilla/5.0 (Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0")
	hash2 := Hash(r, false)
	if expected, got := "2167956e7d64d549b9ede6cb2adcabd5e01194a4", hash2; expected != got {
		t.Fatalf("[2] expected request's hash to be: '%s' but got: '%s'", expected, got)
	}

	if hash1 == hash2 {
		t.Fatalf("expected hashes to not match: %s vs %s", hash1, hash2)
	}

	r.Header.Set("X-something", "something")
	hash3 := Hash(r, false)
	if expected, got := "2fb4398376a47d0f9ba65f85054c4d490d7f466a", hash3; expected != got {
		t.Fatalf("[3] expected request's hash to be: '%s' but got: '%s'", expected, got)
	}

	if hash2 == hash3 {
		t.Fatalf("expected hashes to not match: %s vs %s", hash2, hash3)
	}
}
