package api

import "net/http"

func Api(key string, var1 string, w http.ResponseWriter, r *http.Request) {
	switch key {
	case "user":
		userAPI(var1, w, r)
	case "chat":
		chatAPI(var1, w, r)
	case "file":
		fileAPI(var1, w, r)
	case "messages":
		messagesAPI(var1, w, r)
	}
}
