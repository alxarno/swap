package api2

import (
	"net/http"
	"github.com/Spatium-Messenger/Server/db_api"
	"fmt"
	"encoding/json"
)

func create(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		Name string`json:"name"`
		Type string`json:"type"`
	}
	err:=getJson(&data, r);if err != nil {
		sendAnswerError("failed decode data",0, w);return
	}
	if len(data.Name)<3{
		sendAnswerError("name less then 3 char",0, w);return
	}
	user,err:= testUserToken(data.Token);if err!=nil{
		sendAnswerError("name less then 3 char",0, w);return
	}
	if data.Type == "chat"{
		_,err = db_api.CreateChat(data.Name, user.Id);if err != nil{
			sendAnswerError(err.Error(), 0,w);return
		}
	}
	if data.Type == "channel"{
		_,err = db_api.CreateChannel(data.Name, user.Id);if err != nil{
			sendAnswerError(err.Error(), 0,w);return
		}
	}
	sendAnswerSuccess(w)
}

func addUsers(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string
		Ids []int64
		ChatId int64}
	err:= getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	user,err:= testUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	res,err:=db_api.CheckUserInChatDelete(user.Id, data.ChatId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	if res{sendAnswerError(err.Error(),0, w);return}

	failed := []int64{}
	successAdd := []int64{}
	for i:=0;i<len(data.Ids);i++{
		err:= db_api.InsertUserInChat(data.Ids[i], data.ChatId)
		if err!=nil{
			failed= append(failed, data.Ids[i])
		}else{
			successAdd = append(successAdd, data.Ids[i])
		}
	}
	sendAnswerSuccess(w);
}

func getUsers(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		ChatId int64`json:"chat_id"`}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	_,err= testUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	_,err=testUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	users,err:=db_api.GetChatUserInfo(data.ChatId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	fmt.Fprintf(w, string(users))
	return
}

func deleteUsers(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		Ids []int64`json:"ids"`
		ChatId int64`json:"chat_id"`}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	user,err:= testUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	err=db_api.CheckUserRightsInChat(user.Id,data.ChatId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	err=db_api.DeleteUsersInChat(data.Ids,data.ChatId,false);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	//	Notifications...
	sendAnswerSuccess(w)
}

func recoveryUsers(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		Ids []int64`json:"ids"`
		ChatId int64`json:"chat_id"`}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	user,err:= testUserToken(data.Token);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	err=db_api.CheckUserRightsInChat(user.Id,data.ChatId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	err=db_api.RecoveryUsersInChat(data.Ids,data.ChatId,false);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	//	Notifications...
	sendAnswerSuccess(w)
}

func getChatSettings(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		ChatId int64`json:"chat_id"`}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	user,err:=testUserToken(data.Token);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	err=db_api.CheckUserRightsInChat(user.Id,data.ChatId);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	res,err:=db_api.GetChatSettings(data.ChatId);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	final,err := json.Marshal(res);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	fmt.Fprintf(w, string(final))
}

func setChatSettings(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		ChatId int64`json:"chat_id"`
		Name string`json:"name"`}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	user,err:=testUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	err= db_api.CheckUserRightsInChat(user.Id,data.ChatId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	err= db_api.SetUserSettings(user.Id,data.Name);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
//	Notification
	sendAnswerSuccess(w)
}

func deleteFromDialog(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		ChatId int64`json:"chat_id"`}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	user,err:=testUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	err=db_api.DeleteUsersInChat([]int64{user.Id}, data.ChatId,true);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
//	Notifications...
	sendAnswerSuccess(w)
}

func recoveryUserInDialog(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		ChatId int64`json:"chat_id"`}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	user,err:=testUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	res,err:=db_api.CheckUserInChatDelete(user.Id,data.ChatId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	if !res{
		sendAnswerError("user aren't delete",0, w);return
	}
	err = db_api.RecoveryUsersInChat([]int64{user.Id}, data.ChatId,true);if err!=nil{
		sendAnswerError(err.Error(), 0,w);return
	}
//	Notifications..
	sendAnswerSuccess(w)
}

func deleteChatFromList(w http.ResponseWriter, r *http.Request){
	var data struct{
		Token string`json:"token"`
		ChatId int64`json:"chat_id"`}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("failed decode data",0, w);return
	}
	user,err:=testUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
	err=db_api.DeleteChatFromList(user.Id,data.ChatId);if err!=nil{
		sendAnswerError(err.Error(),0, w);return
	}
//	Notification...
	sendAnswerSuccess(w)
}

func MainChatApi(var1 string, w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch var1 {
	case "create":
		create(w,r)
	case "addUsersInChat":
		addUsers(w,r)
	case "getUsers":
		getUsers(w,r)
	case "deleteUsers":
		deleteUsers(w,r)
	case "recoveryUsers":
		recoveryUsers(w,r)
	case "getSettings":
		getChatSettings(w,r)
	case "setSettings":
		setChatSettings(w,r)
	case "deleteFromDialog":
		deleteFromDialog(w,r)
	case "recoveryUserInDialog":
		recoveryUserInDialog(w,r)
	case "deleteFromList":
		deleteChatFromList(w,r)
	}
}
