package api

import "net/http"

func Api(key string, var1 string, w http.ResponseWriter, r *http.Request) {
	switch key {
	case "user":
		UserApi(var1, w, r)
	case "chat":
		ChatApi(var1, w, r)
	case "file":
		FileApi(var1, w, r)
	case "messages":
		MessagesApi(var1, w, r)
	}
}
