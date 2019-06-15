package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alxarno/swap/models"

	db "github.com/alxarno/swap/db2"
)

func registerChatEndpoints(r *Router) {
	r.Route("/create", create, "POST")
	r.Route("/{id:[0-9]+}/users", getUsers, "GET")
	r.Route("/{id:[0-9]+}/ivite", invite, "POST", "GET")
	r.Route("/{id:[0-9]+}/leave", leaveChat, "POST")
	r.Route("/{id:[0-9]+}/return", turnBackToChat, "POST")
	r.Route("/{id:[0-9]+}/block", blockUsers, "POST")
	r.Route("/{id:[0-9]+}/unblock", recoveryUsers, "POST")
	r.Route("/{id:[0-9]+}/settings", chatSettings, "POST", "GET")
	r.Route("/{id:[0-9]+}/deleteFromList", deleteChatFromList, "POST")
}

func create(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat create API:"
	var data struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	err := getJSON(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	if len(data.Name) < 3 {
		sendAnswerError(ref, nil, data.Name, shortChatName, 1, w)
		return
	}
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 2, w)
		return
	}
	if data.Type == "chat" {
		_, err = db.Create(data.Name, user.ID, db.ChatType)
		if err != nil {
			sendAnswerError(ref, err, fmt.Sprintf("name - %s, userid - %d", data.Name, user.ID), createdChat, 3, w)
			return
		}
	}
	if data.Type == "channel" {
		_, err = db.Create(data.Name, user.ID, db.ChannelType)
		if err != nil {
			sendAnswerError(ref, err, fmt.Sprintf("name - %s, userid - %d", data.Name, user.ID), createdCahnnel, 4, w)
			return
		}
	}
	sendAnswerSuccess(w)
}

func addUsers(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat Add Users API"
	var data struct {
		IDs []int64 `json:"users"`
	}
	err := getJSON(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}

	chatID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
		return
	}
	res, err := db.CheckUserInChatDeleted(user.ID, chatID)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, userid - %d", chatID, user.ID),
			userChatCheckFailed, 2, w)
		return
	}
	if res {
		sendAnswerError(ref, nil, "", userIsDeletedFromChat, 3, w)
		return
	}

	var failed []int64
	var successAdd []int64
	for i := 0; i < len(data.IDs); i++ {
		err := db.InsertUserInChat(data.IDs[i], chatID, true)
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
	_, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
		return
	}
	chatID := pageNumber(r, 2)
	users, err := db.GetChatUsersInfo(chatID)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("chatID - %d", chatID), failedGetUserInfo, 2, w)
		return
	}

	jsondata, err := json.Marshal(users)
	if err != nil {
		return
	}

	fmt.Fprintf(w, string(jsondata))
	return
}

func invite(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUsersForAdd(w, r)
		break
	case http.MethodPost:
		addUsers(w, r)
		break
	default:
		w.WriteHeader(404)
	}
}

func getUsersForAdd(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat get users for add API:"

	var response struct {
		Result string        `json:"result"`
		Users  []models.User `json:"users"`
	}

	_, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, "", invalidToken, 1, w)
		return
	}

	name := r.URL.Query().Get("name")
	chatID := pageNumber(r, 2)

	users := &([]models.User{})

	if len(name) != 0 {
		users, err = db.GetUsersForAddByName(chatID, name)
		if err != nil {
			sendAnswerError(ref, err,
				fmt.Sprintf("chatID - %d, name - %s", chatID, name),
				failedGetUsersForAdd, 4, w)
			return
		}
	}

	var finish []byte
	response.Result = successResult

	if users != nil {
		response.Users = *users
	} else {
		response.Users = []models.User{}
	}
	finish, _ = json.Marshal(response)
	fmt.Fprintf(w, string(finish))
}

func blockUsers(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat delete users API:"
	var data struct {
		IDs []int64 `json:"ids"`
	}
	err := getJSON(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	chatID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
		return
	}
	err = db.CheckUserRights(user.ID, chatID)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, userID - %d", chatID, user.ID),
			haventRightsForAction, 2, w)
		return
	}
	err = db.DeleteUsersInChat(data.IDs, chatID, false)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, IDs - %d", chatID, data.IDs),
			failedDeleteUsers, 3, w)
		return
	}
	sendAnswerSuccess(w)
}

func leaveChat(w http.ResponseWriter, r *http.Request) {
	const ref string = "Leave chat API:"
	chatID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
		return
	}
	err = db.DeleteUsersInChat([]int64{user.ID}, chatID, true)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, ID - %d", chatID, user.ID),
			failedDeleteUsers, 4, w)
		return
	}
	sendAnswerSuccess(w)
}

func turnBackToChat(w http.ResponseWriter, r *http.Request) {
	const ref string = "Turn Back to Chat API:"
	chatID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
		return
	}
	err = db.RecoveryUsersInChat([]int64{user.ID}, chatID, true)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, ID - %d", chatID, user.ID),
			failedDeleteUsers, 4, w)
		return
	}
	sendAnswerSuccess(w)
}

func recoveryUsers(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat recovery users API:"
	var data struct {
		IDs []int64 `json:"ids"`
	}
	err := getJSON(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	chatID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
		return
	}
	err = db.CheckUserRights(user.ID, chatID)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, user id - %d", chatID, user.ID),
			haventRightsForAction, 2, w)
		return
	}
	err = db.RecoveryUsersInChat(data.IDs, chatID, false)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, IDs - %d", chatID, data.IDs),
			failedRecoveryUsers, 3, w)
		return
	}
	//	Notifications...
	sendAnswerSuccess(w)
}

func getChatSettings(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat get chat settings API:"
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, "", invalidToken, 1, w)
		return
	}
	chatID := pageNumber(r, 2)

	err = db.CheckUserRights(user.ID, chatID)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, userID - %d", chatID, user.ID),
			haventRightsForAction, 2, w)
		return
	}
	res, err := db.GetChatSettings(chatID)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d", chatID),
			failedGetChatSettings, 3, w)
		return
	}
	final, err := json.Marshal(res)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("data - %v", res),
			failedEncodeData, 3, w)
		return
	}
	fmt.Fprintf(w, string(final))
}

func setChatSettings(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat set chat settings API:"
	var data struct {
		Name string `json:"name"`
	}

	err := getJSON(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	chatID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, "", invalidToken, 1, w)
		return
	}
	// log.Println(user)
	err = db.CheckUserRights(user.ID, chatID)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, userID - %d", chatID, user.ID),
			haventRightsForAction, 2, w)
		return
	}
	err = db.SetChatSettings(chatID, models.ChatSettings{Name: data.Name})
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, name - %s", chatID, data.Name),
			haventRightsForAction, 3, w)
		return
	}
	//	Notification
	sendAnswerSuccess(w)
}

func chatSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getChatSettings(w, r)
		break
	case http.MethodPost:
		setChatSettings(w, r)
		break
	default:
		w.WriteHeader(404)
	}
}

func deleteChatFromList(w http.ResponseWriter, r *http.Request) {
	const ref string = "Chat delete chat from list API:"
	chatID := pageNumber(r, 2)
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
		return
	}
	err = db.DeleteChatFromList(user.ID, chatID)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("chatID - %d, userID - %d", chatID, user.ID),
			failedDeleteFromList, 2, w)
		return
	}
	sendAnswerSuccess(w)
}

// type dataTokenChat struct {
// 	// Token  string `json:"token"`
// 	ChatID int64 `json:"chat_id,integer"`
// }

// func deleteFromDialog(w http.ResponseWriter, r *http.Request) {
// 	const ref string = "Chat delete from dialog API:"
// 	var data dataTokenChat

// 	err := getJSON(&data, r)
// 	if err != nil {
// 		decodeFail(ref, err, r, w)
// 		return
// 	}
// 	user, err := UserByHeader(r)
// 	if err != nil {
// 		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
// 		return
// 	}
// 	err = db.DeleteUsersInChat([]int64{user.ID}, data.ChatID, true)
// 	if err != nil {
// 		sendAnswerError(ref, err,
// 			fmt.Sprintf("chatID - %d, userID - %d", data.ChatID, user.ID),
// 			failedDeleteUsers, 2, w)
// 		return
// 	}
// 	//	Notifications...
// 	sendAnswerSuccess(w)
// }

// func recoveryUserInDialog(w http.ResponseWriter, r *http.Request) {
// 	const ref string = "Chat recovery user in dialog API:"
// 	var data dataTokenChat

// 	err := getJSON(&data, r)
// 	if err != nil {
// 		decodeFail(ref, err, r, w)
// 		return
// 	}
// 	user, err := UserByHeader(r)
// 	if err != nil {
// 		sendAnswerError(ref, err, getToken(r), invalidToken, 1, w)
// 		return
// 	}
// 	res, err := db.CheckUserInChatDeleted(user.ID, data.ChatID)
// 	if err != nil {
// 		sendAnswerError(ref, err,
// 			fmt.Sprintf("chatID - %d, userID - %d", data.ChatID, user.ID),
// 			userChatCheckFailed, 2, w)
// 		return
// 	}
// 	if !res {
// 		sendAnswerError(ref, err,
// 			fmt.Sprintf("chatID - %d, userID - %d", data.ChatID, user.ID),
// 			userIsDeletedFromChat, 2, w)
// 		return
// 	}
// 	err = db.RecoveryUsersInChat([]int64{user.ID}, data.ChatID, true)
// 	if err != nil {
// 		sendAnswerError(ref, err,
// 			fmt.Sprintf("chatID - %d, userID - %d", data.ChatID, user.ID),
// 			failedRecoveryUsers, 2, w)
// 		return
// 	}
// 	sendAnswerSuccess(w)
// }

// func usersForDialog(w http.ResponseWriter, r *http.Request) {
// 	const ref string = "Chat users for dialog API:"
// 	var data struct {
// 		Token string `json:"token"`
// 		Name  string `json:"name"`
// 	}
// 	err := getJSON(&data, r)
// 	if err != nil {
// 		decodeFail(ref, err, r, w)
// 		return
// 	}
// 	user, err := TestUserToken(data.Token)
// 	if err != nil {
// 		sendAnswerError(ref, err, data.Token, invalidToken, 1, w)
// 		return
// 	}
// 	users, err := db.GetUsersForCreateDialog(user.ID, data.Name)
// 	if err != nil {
// 		sendAnswerError(ref, err,
// 			fmt.Sprintf("userID - %d, name - %s", user.ID, data.Name),
// 			failedGetUsersForDialog, 2, w)
// 		return
// 	}
// 	finish, _ := json.Marshal(users)
// 	fmt.Fprintf(w, string(finish))
// }
