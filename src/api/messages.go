package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	db "github.com/alxarno/swap/db2"
	"github.com/alxarno/swap/models"
)

func registerMessagesEndpoints(r *Router) {
	r.Route("/", getMessages, "GET")
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	const ref string = "Messages get API:"
	chatID, err := strconv.ParseInt(r.URL.Query().Get("chat"), 10, 64)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), failedDecodeData, 0, w)
		return
	}
	var last int64
	if r.URL.Query().Get("last") != "" {
		last, err = strconv.ParseInt(r.URL.Query().Get("last"), 10, 64)
		if err != nil {
			sendAnswerError(ref, err, getToken(r), failedDecodeData, 1, w)
			return
		}
	}

	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 2, w)
		return
	}

	//There is no need check user is in chat, because func "GetMessage" check it auto
	var mes *[]models.NewMessageToUser
	if last == 0 {
		mes, err = db.GetMessages(user.ID, chatID, false, 0)
		if err != nil {
			sendAnswerError(ref, err, fmt.Sprintf("userID - %d, chatID - %d", user.ID, chatID), failedGetMessages, 3, w)
			return
		}
	} else {
		mes, err = db.GetMessages(user.ID, chatID, true, last)
		if err != nil {
			sendAnswerError(ref, err, fmt.Sprintf("userID - %d, chatID - %d", user.ID, chatID), failedGetAdditionalMessages, 4, w)
			return
		}
	}

	var finish []byte
	if mes == nil {
		finish, _ = json.Marshal([]string{})
	} else {
		finish, _ = json.Marshal(*mes)
	}
	fmt.Fprintf(w, string(finish))
}
