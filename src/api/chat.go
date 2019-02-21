package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/swap-messenger/Backend/db"
)

func create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string `json:"token"`
		Name  string `json:"name"`
		Type  string `json:"type"`
	}
	err := getJson(&data, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	if len(data.Name) < 3 {
		sendAnswerError("name less then 3 char", 1, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError("token is invalid", 2, w)
		return
	}
	if data.Type == "chat" {
		_, err = db.CreateChat(data.Name, user.Id)
		if err != nil {
			sendAnswerError(err.Error(), 3, w)
			return
		}
	}
	if data.Type == "channel" {
		_, err = db.CreateChannel(data.Name, user.Id)
		if err != nil {
			sendAnswerError(err.Error(), 4, w)
			return
		}
	}
	sendAnswerSuccess(w)
}

func addUsers(w http.ResponseWriter, r *http.Request) {
	var sData struct {
		Token  string    `json:"token"`
		Ids    []float64 `json:"users"`
		ChatId float64   `json:"chat_id"`
	}
	var data struct {
		Token  string
		Ids    []int64
		ChatId int64
	}
	err := getJson(&sData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(sData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 1, w)
		return
	}
	res, err := db.CheckUserInChatDelete(user.Id, data.ChatId)
	if err != nil {
		sendAnswerError(err.Error(), 2, w)
		return
	}
	if res {
		sendAnswerError(err.Error(), 3, w)
		return
	}

	var failed []int64
	var successAdd []int64
	for i := 0; i < len(data.Ids); i++ {
		err := db.InsertUserInChat(data.Ids[i], data.ChatId, true)
		if err != nil {
			failed = append(failed, data.Ids[i])
		} else {
			successAdd = append(successAdd, data.Ids[i])
		}
	}
	sendAnswerSuccess(w)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var pData struct {
		Token  string  `json:"token"`
		ChatId float64 `json:"chat_id"`
	}
	var data struct {
		Token  string
		ChatId int64
	}
	err := getJson(&pData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(pData, &data)
	_, err = TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 1, w)
		return
	}
	users, err := db.GetChatUserInfo(data.ChatId)
	if err != nil {
		sendAnswerError(err.Error(), 2, w)
		return
	}
	fmt.Fprintf(w, string(users))
	return
}

func getUsersForAdd(w http.ResponseWriter, r *http.Request) {
	var pData struct {
		Token  string  `json:"token"`
		ChatId float64 `json:"chat_id"`
		Name   string  `json:"name"`
	}
	var data struct {
		Token  string
		ChatId int64
		Name   string
	}
	err := getJson(&pData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(pData, &data)
	_, err = TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 1, w)
		return
	}
	users, err := db.GetUsersForAddByName(data.ChatId, data.Name)
	if err != nil {
		sendAnswerError(err.Error(), 2, w)
		return
	}
	var finish []byte
	var x = make(map[string]interface{})
	x["result"] = "Success"
	if users == nil {
		x["users"] = [0]map[string]interface{}{}
	} else {
		x["users"] = users
	}
	finish, _ = json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func deleteUsers(w http.ResponseWriter, r *http.Request) {
	var rData struct {
		Token  string    `json:"token"`
		Ids    []float64 `json:"ids"`
		ChatId float64   `json:"chat_id"`
	}
	var data struct {
		Token  string  `json:"token"`
		Ids    []int64 `json:"ids"`
		ChatId int64   `json:"chat_id"`
	}
	err := getJson(&rData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(rData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	err = db.CheckUserRightsInChat(user.Id, data.ChatId)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	err = db.DeleteUsersInChat(data.Ids, data.ChatId, false)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	//	Notifications...
	sendAnswerSuccess(w)
}

func recoveryUsers(w http.ResponseWriter, r *http.Request) {
	var rData struct {
		Token  string    `json:"token"`
		Ids    []float64 `json:"ids"`
		ChatId float64   `json:"chat_id"`
	}
	var data struct {
		Token  string  `json:"token"`
		Ids    []int64 `json:"ids"`
		ChatId int64   `json:"chat_id"`
	}
	err := getJson(&rData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(rData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	err = db.CheckUserRightsInChat(user.Id, data.ChatId)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	err = db.RecoveryUsersInChat(data.Ids, data.ChatId, false)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	//	Notifications...
	sendAnswerSuccess(w)
}

func getChatSettings(w http.ResponseWriter, r *http.Request) {
	var rData struct {
		Token  string  `json:"token"`
		ChatId float64 `json:"chat_id"`
	}
	var data struct {
		Token  string `json:"token"`
		ChatId int64  `json:"chat_id"`
	}
	err := getJson(&rData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(rData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	err = db.CheckUserRightsInChat(user.Id, data.ChatId)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	res, err := db.GetChatSettings(data.ChatId)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	final, err := json.Marshal(res)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	fmt.Fprintf(w, string(final))
}

func setChatSettings(w http.ResponseWriter, r *http.Request) {
	var rData struct {
		Token  string  `json:"token"`
		ChatId float64 `json:"chat_id"`
		Name   string  `json:"name"`
	}
	var data struct {
		Token  string `json:"token"`
		ChatId int64  `json:"chat_id"`
		Name   string `json:"name"`
	}
	err := getJson(&rData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(rData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	err = db.CheckUserRightsInChat(user.Id, data.ChatId)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	err = db.SetUserSettings(user.Id, data.Name)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	//	Notification
	sendAnswerSuccess(w)
}

func deleteFromDialog(w http.ResponseWriter, r *http.Request) {
	var rData struct {
		Token  string  `json:"token"`
		ChatId float64 `json:"chat_id"`
	}
	var data struct {
		Token  string `json:"token"`
		ChatId int64  `json:"chat_id"`
	}
	err := getJson(&rData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(rData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	err = db.DeleteUsersInChat([]int64{user.Id}, data.ChatId, true)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	//	Notifications...
	sendAnswerSuccess(w)
}

func recoveryUserInDialog(w http.ResponseWriter, r *http.Request) {
	var rData struct {
		Token  string  `json:"token"`
		ChatId float64 `json:"chat_id"`
	}
	var data struct {
		Token  string `json:"token"`
		ChatId int64  `json:"chat_id"`
	}
	err := getJson(&rData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(rData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	res, err := db.CheckUserInChatDelete(user.Id, data.ChatId)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	if !res {
		sendAnswerError("user aren't delete", 0, w)
		return
	}
	err = db.RecoveryUsersInChat([]int64{user.Id}, data.ChatId, true)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	//	Notifications..
	sendAnswerSuccess(w)
}

func deleteChatFromList(w http.ResponseWriter, r *http.Request) {
	var rData struct {
		Token  string  `json:"token"`
		ChatId float64 `json:"chat_id"`
	}
	var data struct {
		Token  string `json:"token"`
		ChatId int64  `json:"chat_id"`
	}
	err := getJson(&rData, r)
	if err != nil {
		sendAnswerError("failed decode data", 0, w)
		return
	}
	TypeChanger(rData, &data)
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	err = db.DeleteChatFromList(user.Id, data.ChatId)
	if err != nil {
		sendAnswerError(err.Error(), 0, w)
		return
	}
	//	Notification...
	sendAnswerSuccess(w)
}

func ChatApi(var1 string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch var1 {
	case "create":
		create(w, r)
	case "addUsersInChat":
		addUsers(w, r)
	case "getUsersForAdd":
		getUsersForAdd(w, r)
	case "getUsers":
		getUsers(w, r)
	case "deleteUsers":
		deleteUsers(w, r)
	case "recoveryUsers":
		recoveryUsers(w, r)
	case "getSettings":
		getChatSettings(w, r)
	case "setSettings":
		setChatSettings(w, r)
	case "deleteFromDialog":
		deleteFromDialog(w, r)
	case "recoveryUserInDialog":
		recoveryUserInDialog(w, r)
	case "deleteFromList":
		deleteChatFromList(w, r)
	default:
		sendAnswerError("Not found", 0, w)
	}
}
