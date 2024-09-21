package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NamedHandleFunc registers a route with a given name.
func NamedHandleFunc(router *mux.Router, path string, handler http.HandlerFunc, methods []string, name string) {
    router.HandleFunc(path, handler).Methods(methods...).Name(name)
}