package handler

import (
	"context"
	"net/http"
)

// Handler is an http.Handler interface with a more expressive function
// signature that expects a byte array to be returned
type Handler func(context.Context, http.ResponseWriter, *http.Request) (
	interface{}, error)

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := h(r.Context(), w, r)
	rawResponse(w, b, err)
}

type HandlerFunc func(Handler) Handler

// JSON is an http.handler interface that ensures that all responses, including
// unexpected errors, are returned in a consistent JSON format. JSON is also
// extended in middleware.go to allow easy chaining of JSON middleware
type JSON Handler

func (h JSON) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := h(r.Context(), w, r)
	jsonResponse(w, b, err)
}
