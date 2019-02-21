package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/swap-messenger/Backend/db"
	"github.com/swap-messenger/Backend/models"
	//"github.com/AlexeyArno/Gologer"
)

func getMessages(w http.ResponseWriter, r *http.Request) {
	var rData struct {
		Token  string  `json:"token"`
		LastId float64 `json:"last_index"`
		ChatId float64 `json:"chat_id"`
	}
	var data struct {
		Token  string `json:"token"`
		LastId int64  `json:"last_index"`
		ChatId int64  `json:"chat_id"`
	}
	err := getJson(&rData, r)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	TypeChanger(rData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 1, w)
		return
	}

	//There is no need check user is in chat, because func "GetMessage" check it auto
	var mes []*models.NewMessageToUser
	if data.LastId == 0 {
		mes, err = db.GetMessages(user.Id, data.ChatId, false, 0)
		if err != nil {
			sendAnswerError(err.Error(), 2, w)
			return
		}
	} else {
		mes, err = db.GetMessages(user.Id, data.ChatId, true, data.LastId)
		if err != nil {
			sendAnswerError(err.Error(), 3, w)
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
		sendAnswerError("Not found", 0, w)
	}
}
