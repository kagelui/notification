package web

import (
	"net/http"
)

// HandlerFunc is a http.HandlerFunc variant that returns error
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// HandlerWrapper is the function signature to chain HandlerFunc
type HandlerWrapper func(next HandlerFunc) HandlerFunc

// Wrap chains the provided list of HandlerWrapper to the provided HandlerFunc
func Wrap(h HandlerFunc, adapters ...HandlerWrapper) HandlerFunc {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

// Handler is a http.Handler implementation that handles HandlerFunc
type Handler struct {
	H HandlerFunc
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.H(w, r); err != nil {
		RespondJSON(r.Context(), w, err, nil)
	}
}
