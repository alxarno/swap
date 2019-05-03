package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/swap-messenger/swap/db"
)

type dataTokenChat struct {
	Token  string `json:"token"`
	ChatID int64  `json:"chat_id,integer"`
}

func create(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat create API:"
	var data struct {
		Token string `json:"token"`
		Name  string `json:"name"`
		Type  string `json:"type"`
	}
	err := getJson(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	if len(data.Name) < 3 {
		sendAnswerError(ref, nil, data.Name, SHORT_CHAT_NAME, 1, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(ref, err, data.Token, INVALID_TOKEN, 2, w)
		return
	}
	if data.Type == "chat" {
		_, err = db.CreateChat(data.Name, user.Id)
		if err != nil {
			sendAnswerError(ref, err, map[string]interface{}{"name": data.Name, "userID": user.Id}, CREATED_CHAT, 3, w)
			return
		}
	}
	if data.Type == "channel" {
		_, err = db.CreateChannel(data.Name, user.Id)
		if err != nil {
			sendAnswerError(ref, err, map[string]interface{}{"name": data.Name, "userID": user.Id}, CREATED_CHANNEL, 4, w)
			log.Println("Chat create API: 4 - ", err.Error())
			return
		}
	}
	sendAnswerSuccess(w)
}

func addUsers(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat Add Users API"
	var data struct {
		Token  string  `json:"token"`
		IDs    []int64 `json:"users"`
		ChatID int64   `json:"chat_id,integer"`
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
	res, err := db.CheckUserInChatDelete(user.Id, data.ChatID)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "userID": user.Id},
			USER_CHAT_CHECK_FAILED, 2, w)
		return
	}
	if res {
		sendAnswerError(ref, nil, nil, USER_IS_DELETED_FROM_CHAT, 3, w)
		return
	}

	var failed []int64
	var successAdd []int64
	for i := 0; i < len(data.IDs); i++ {
		err := db.InsertUserInChat(data.IDs[i], data.ChatID, true)
		if err != nil {
			failed = append(failed, data.IDs[i])
		} else {
			successAdd = append(successAdd, data.IDs[i])
		}
	}
	sendAnswerSuccess(w)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat get users API:"
	var data dataTokenChat

	err := getJson(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	_, err = TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(ref, err, data.Token, INVALID_TOKEN, 1, w)
		return
	}
	users, err := db.GetChatUserInfo(data.ChatID)
	if err != nil {
		sendAnswerError(ref, err, map[string]interface{}{"chatID": data.ChatID}, FAILED_GET_USER_INFO, 2, w)
		return
	}
	fmt.Fprintf(w, string(users))
	return
}

func getUsersForAdd(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat get users for add API:"
	var data struct {
		Token  string `json:"token"`
		ChatID int64  `json:"chat_id,integer"`
		Name   string `json:"name"`
	}

	err := getJson(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	_, err = TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(ref, err, data.Token, INVALID_TOKEN, 1, w)
		return
	}
	users, err := db.GetUsersForAddByName(data.ChatID, data.Name)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "name": data.Name},
			FAILED_GET_USERS_FOR_ADD, 2, w)
		return
	}
	var finish []byte
	var x = make(map[string]interface{})
	x["result"] = SUCCESS_ANSWER
	if users == nil {
		x["users"] = [0]map[string]interface{}{}
	} else {
		x["users"] = users
	}
	finish, _ = json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func deleteUsers(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat delete users API:"
	var data struct {
		Token  string  `json:"token"`
		IDs    []int64 `json:"ids"`
		ChatID int64   `json:"chat_id,integer"`
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
	err = db.CheckUserRightsInChat(user.Id, data.ChatID)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			HAVENT_RIGHTS_FOR_ACTION, 2, w)
		return
	}
	err = db.DeleteUsersInChat(data.IDs, data.ChatID, false)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "IDs": data.IDs},
			FAILED_DELETE_USERS, 3, w)
		return
	}
	sendAnswerSuccess(w)
}

func recoveryUsers(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat recovery users API:"
	var data struct {
		Token  string  `json:"token"`
		IDs    []int64 `json:"ids"`
		ChatID int64   `json:"chat_id,integer"`
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
	err = db.CheckUserRightsInChat(user.Id, data.ChatID)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			HAVENT_RIGHTS_FOR_ACTION, 2, w)
		return
	}
	err = db.RecoveryUsersInChat(data.IDs, data.ChatID, false)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "ids": data.IDs},
			FAILED_RECOVERY_USERS, 3, w)
		return
	}
	//	Notifications...
	sendAnswerSuccess(w)
}

func getChatSettings(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat get chat settings API:"
	var data dataTokenChat

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
	err = db.CheckUserRightsInChat(user.Id, data.ChatID)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			HAVENT_RIGHTS_FOR_ACTION, 2, w)
		return
	}
	res, err := db.GetChatSettings(data.ChatID)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID},
			FAILED_GET_CHAT_SETTINGS, 3, w)
		return
	}
	final, err := json.Marshal(res)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"data": res},
			FAILED_ENCODE_DATA, 3, w)
		return
	}
	fmt.Fprintf(w, string(final))
}

func setChatSettings(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat set chat settings API:"
	var data struct {
		Token  string `json:"token"`
		ChatID int64  `json:"chat_id"`
		Name   string `json:"name"`
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
	// log.Println(user)
	err = db.CheckUserRightsInChat(user.Id, data.ChatID)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			HAVENT_RIGHTS_FOR_ACTION, 2, w)
		return
	}
	err = db.SetNameChat(data.ChatID, data.Name)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "name": data.Name},
			HAVENT_RIGHTS_FOR_ACTION, 3, w)
		return
	}
	//	Notification
	sendAnswerSuccess(w)
}

func deleteFromDialog(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat delete from dialog API:"
	var data dataTokenChat

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
	err = db.DeleteUsersInChat([]int64{user.Id}, data.ChatID, true)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			FAILED_DELETE_USERS, 2, w)
		return
	}
	//	Notifications...
	sendAnswerSuccess(w)
}

func recoveryUserInDialog(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat recovery user in dialog API:"
	var data dataTokenChat

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
	res, err := db.CheckUserInChatDelete(user.Id, data.ChatID)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			USER_CHAT_CHECK_FAILED, 2, w)
		return
	}
	if !res {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			USER_IS_DELETED_FROM_CHAT, 2, w)
		return
	}
	err = db.RecoveryUsersInChat([]int64{user.Id}, data.ChatID, true)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			FAILED_RECOVERY_USERS, 2, w)
		return
	}
	//	Notifications..
	sendAnswerSuccess(w)
}

func deleteChatFromList(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat delete chat from list API:"
	var data dataTokenChat

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
	err = db.DeleteChatFromList(user.Id, data.ChatID)
	if err != nil {
		sendAnswerError(ref, err,
			map[string]interface{}{"chatID": data.ChatID, "id": user.Id},
			FAILED_DELETE_FROM_LIST, 2, w)
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
		sendAnswerError("Chat API Router", nil, nil, END_POINT_NOT_FOUND, 0, w)
	}
}
