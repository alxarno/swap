package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	db "github.com/alxarno/swap/db2"
	"github.com/alxarno/swap/models"
)

func getMessages(w *http.ResponseWriter, r *http.Request) {
	const ref string = "Messages get API:"
	var data struct {
		LastID int64 `json:"last_index,integer"`
		ChatID int64 `json:"chat_id,integer"`
	}

	err := getJSON(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	user, err := getUserByToken(r)
	if err != nil {
		sendAnswerError(ref, err, r.Header.Get("X-Auth-Token"), invalidToken, 1, w)
		return
	}

	//There is no need check user is in chat, because func "GetMessage" check it auto
	var mes *[]models.NewMessageToUser
	if data.LastID == 0 {
		mes, err = db.GetMessages(user.ID, data.ChatID, false, 0)
		if err != nil {
			sendAnswerError(ref, err, fmt.Sprintf("userID - %d, chatID - %d", user.ID, data.ChatID), failedGetMessages, 2, w)
			return
		}
	} else {
		mes, err = db.GetMessages(user.ID, data.ChatID, true, data.LastID)
		if err != nil {
			sendAnswerError(ref, err, fmt.Sprintf("userID - %d, chatID - %d", user.ID, data.ChatID), failedGetAdditionalMessages, 3, w)
			return
		}
	}

	var finish []byte
	if mes == nil {
		finish, _ = json.Marshal([]string{})
	} else {
		finish, _ = json.Marshal(*mes)
	}
	fmt.Fprintf((*w), string(finish))
}

func messagesAPI(var1 string, w *http.ResponseWriter, r *http.Request) {
	switch var1 {
	case "messages":
		getMessages(w, r)
	default:
		sendAnswerError("Messages API Router", nil, "", endPointNotFound, 0, w)
	}
}
