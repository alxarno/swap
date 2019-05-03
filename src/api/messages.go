package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/swap-messenger/swap/db"
	"github.com/swap-messenger/swap/models"
	//"github.com/AlexeyArno/Gologer"
)

func getMessages(w http.ResponseWriter, r *http.Request) {
	const ref string = "Messages get API:"
	var data struct {
		Token  string `json:"token"`
		LastID int64  `json:"last_index,integer"`
		ChatID int64  `json:"chat_id,integer"`
	}

	err := getJson(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(ref, err, data.Token, INVALID_TOKEN, 1, w)
		return
	}

	//There is no need check user is in chat, because func "GetMessage" check it auto
	var mes []*models.NewMessageToUser
	if data.LastID == 0 {
		mes, err = db.GetMessages(user.Id, data.ChatID, false, 0)
		if err != nil {
			sendAnswerError(ref, err, map[string]interface{}{"userID": user.Id, "chatID": data.ChatID}, FAILED_GET_MESSAGES, 2, w)
			return
		}
	} else {
		mes, err = db.GetMessages(user.Id, data.ChatID, true, data.LastID)
		if err != nil {
			sendAnswerError(ref, err, map[string]interface{}{"userID": user.Id, "chatID": data.ChatID}, FAILED_GET_ADDITIONAL_MESSAGES, 3, w)
			return
		}
	}

	var finish []byte
	if mes == nil {
		finish, _ = json.Marshal([]string{})
	} else {
		finish, _ = json.Marshal(mes)
	}
	fmt.Fprintf(w, string(finish))
}

//func getEarlyMessages(w http.ResponseWriter, r *http.Request){
//	var data struct{
//		ChatId int64`json:"chat_id"`
//		LastId int64`json:"last_id"`
//		Token string`json:"token"`
//	}
//	err:=getJson(&data,r);if err!=nil{
//		sendAnswerError(err.Error(),0,w);return
//	}
//	user,err:=TestUserToken(data.Token);if err!=nil{
//		sendAnswerError(err.Error(),0,w);return
//	}
//	_,err=db.CheckUserInChatDelete(user.Id, data.ChatId);if err!=nil{
//		sendAnswerError(err.Error(),0,w);return
//	}
//	mes,err:=db.GetMessages(user.Id,data.ChatId,true,data.LastId);if err!=nil{
//		sendAnswerError(err.Error(),0,w);return
//	}
//	var finish []byte
//	if mes == nil{
//		finish, _=json.Marshal([]string{})
//	}else{
//		finish, _=json.Marshal(mes)
//	}
//	fmt.Fprintf(w, string(finish))
//}

func MessagesApi(var1 string, w http.ResponseWriter, r *http.Request) {
	switch var1 {
	case "getMessages":
		getMessages(w, r)
	default:
		sendAnswerError("Messages API Router", nil, nil, END_POINT_NOT_FOUND, 0, w)
	}
}
