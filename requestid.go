package requestid

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/google/uuid"
)

const xRequestIDHeaderKey = "X-Request-Id"

// Generator defines the function which should extract or generate
// a Request ID. See `DefaultGenerator` package-level function.
type Generator func(w http.ResponseWriter, r *http.Request) string

// DefaultGenerator is the default `Generator`.
// It extracts the ID from the "X-Request-Id" request header value
// or, if missing, it generates a new UUID(v4) and sets the response header.
//
// See `Get` package-level function too.
var DefaultGenerator Generator = func(w http.ResponseWriter, r *http.Request) string {
	id := w.Header().Get(xRequestIDHeaderKey) // already set by prior middleware.
	if id != "" {
		return id
	}

	id = r.Header.Get(xRequestIDHeaderKey)
	if id == "" {
		uid, err := uuid.NewRandom()
		if err != nil {
			return ""
		}

		id = uid.String()
	}

	setHeader(w, id)
	return id
}

// HashGenerator uses the request's hash to generate a fixed-length Request ID.
// Note that one or many requests may contain the same ID, so it's not unique.
func HashGenerator(includeBody bool) Generator {
	return func(w http.ResponseWriter, r *http.Request) string {
		w.Header().Set(xRequestIDHeaderKey, Hash(r, includeBody))
		return DefaultGenerator(w, r)
	}
}

// ErrorHandler is the handler that is executed when a Generator
// returns an empty string.
var ErrorHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
})

// Handler wraps a handler with the requestid middleware
// using the `DefaultGenerator`.
// See `Get` package-level function to
// retrieve the generated-and-stored request id.
// See `HandlerWithGenerator` to use a custom request ID generator.
func Handler(next http.Handler) http.HandlerFunc {
	return HandlerWithGenerator(next, DefaultGenerator)
}

// HandlerWithGenerator same as `Handler` function
// but it accepts a custom `Generator`
// to extract (and set) the request ID.
func HandlerWithGenerator(next http.Handler, gen Generator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := gen(w, r)
		if id == "" {
			ErrorHandler.ServeHTTP(w, r)
			return
		}

		r = Set(r, id)
		next.ServeHTTP(w, r)
	}
}

var requestIDContextKey interface{} = "requestid"

// Set manually sets a Request ID for this request.
// Returns the shallow copy of given "r" request
// contains the new ID context value.
// Can be called before Handler execution to modify the
// method of extraction of the ID.
//
// Note: Caller should manually set a response header for the client, if necessary.
//
// See `Get` package-level function too.
func Set(r *http.Request, id string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), requestIDContextKey, id))
}

// Get returns the Request ID of this request.
// A prior call to the Handler or HandlerWithGenerator is required.
func Get(r *http.Request) string {
	v := r.Context().Value(requestIDContextKey)
	if v != nil {
		if id, ok := v.(string); ok {
			return id
		}
	}

	return ""
}

func setHeader(w http.ResponseWriter, id string) {
	w.Header().Set(xRequestIDHeaderKey, id)
}

// Hash returns the sha1 hash of the "r" request.
// It does not capture error, instead it returns an empty string.
func Hash(r *http.Request, includeBody bool) string {
	h := sha1.New()
	b, err := httputil.DumpRequest(r, includeBody)
	if err != nil {
		return ""
	}
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
