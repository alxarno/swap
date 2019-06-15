package api

// RegisterEndpoints - registr endpoints
func RegisterEndpoints(r *Router) {
	registerUserEndpoints(r.Subroute("/user"))
	registerChatEndpoints(r.Subroute("/chat"))
	registerFileEndpoints(r.Subroute("/file"))
	registerMessagesEndpoints(r.Subroute("/messages"))
}
