package api

import "net/http"

type Router struct {
	Route    func(pattern string, handler func(w http.ResponseWriter, r *http.Request), methods ...string)
	Subroute func(pattern string) *Router
}
