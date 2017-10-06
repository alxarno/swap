package api

import (
	"net/http"
	chat "github.com/AlexeyArno/Spatium/src/api/chat"
	user "github.com/AlexeyArno/Spatium/src/api/user"

)



func MainApiRouter(key string, var1 string, w http.ResponseWriter, r *http.Request){
	switch key {
		case "user":
			user.MainUserApi(var1, w , r )
		case "chat":
			chat.MainChatApi(var1, w , r )
		}
}
