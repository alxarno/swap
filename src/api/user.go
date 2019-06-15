package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/alxarno/swap/logger"

	"github.com/alxarno/swap/models"

	db "github.com/alxarno/swap/db2"
)

type onlineUsers = func(users *[]int64) int64

var (
	// GetOnlineUsers - external function in 'messages'  for getting online users count
	GetOnlineUsers onlineUsers
)

func registerUserEndpoints(r *Router) {
	r.Route("/enter", enter, "POST")
	r.Route("/tokencheck", proveToken, "GET")
	r.Route("/create", createUser, "POST")
	r.Route("/chats", userChats, "GET")
	r.Route("/data", userData, "GET")
	r.Route("/settings", setSettings, "POST")
}

func enter(w http.ResponseWriter, r *http.Request) {
	const ref string = "User enter API:"
	var data struct {
		Login string `json:"login"`
		Pass  string `json:"pass"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	user, err := db.GetUserByLoginAndPass(data.Login, data.Pass)
	if err != nil {
		sendAnswerError(ref, err,
			fmt.Sprintf("login - %s, pass - %s", data.Login, strings.Repeat("*", utf8.RuneCountInString(data.Pass))),
			failedGetUser, 1, w)
		return
	}

	//if user.CheckPass(data.Pass){
	token, err := generateToken(user.ID)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("userID - %d", user.ID), failedGenerateToken, 2, w)
		return
	}
	var x = make(map[string]string)
	x["token"] = token
	x["result"] = successResult
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))
	return
}

func proveToken(w http.ResponseWriter, r *http.Request) {
	const ref string = "User proveToken API:"
	var x = struct {
		Result string `json:"result"`
		Code   int64  `json:"code"`
	}{
		Code:   0,
		Result: successResult,
	}
	_, err := UserByHeader(r)
	if err != nil {
		x.Result = errorResult
	}
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func createUser(w http.ResponseWriter, r *http.Request) {
	const ref string = "User create user API:"
	var data struct {
		Login string `json:"login"`
		Pass  string `json:"pass"`
	}
	err := getJSON(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}

	passPrint := strings.Repeat("*", utf8.RuneCountInString(data.Pass))
	if data.Login == "" || data.Pass == "" {
		sendAnswerError(ref, err, fmt.Sprintf("login - %s, pass - %s", data.Login, passPrint), someEmptyFields, 1, w)
		return
	}
	id, err := db.CreateUser(data.Login, data.Pass, data.Login)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("login - %s, pass - %s", data.Login, passPrint), failedCreateUser, 2, w)
		return
	}
	token, err := generateToken(id)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("userID - %d", id), failedGenerateToken, 3, w)
		return
	}
	var x = struct {
		Result string `json:"result"`
		Token  string `json:"token"`
	}{
		Token:  token,
		Result: successResult,
	}
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))
	return
}

func userChats(w http.ResponseWriter, r *http.Request) {
	const ref string = "User get chats API:"
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, "", invalidToken, 1, w)
		return
	}
	chats, err := db.GetUserChats(user.ID)
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("userID - %d", user.ID), failedGetUserChats, 2, w)
		return
	}
	var finish []byte
	if chats == nil {
		finish = []byte("[]")
	} else {
		if GetOnlineUsers != nil {
			for i, v := range *chats {
				chatUsers, err := db.GetChatsUsers(v.ID)
				if err != nil {
					logger.Logger.Println(fmt.Sprintf("%s Cant get chat users -> %s", ref, err))
					continue
				}

				v.Online = GetOnlineUsers(chatUsers)
				(*chats)[i] = v
			}
		}
		finish, _ = json.Marshal(*chats)
	}
	fmt.Fprintf(w, string(finish))
}

func userData(w http.ResponseWriter, r *http.Request) {
	const ref string = "User get data API:"
	user, err := UserByHeader(r)
	if err != nil {
		sendAnswerError(ref, err, "", invalidToken, 1, w)
		return
	}
	data := struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Language string `json:"language"`
	}{
		ID:       user.ID,
		Language: user.Language,
		Name:     user.Name,
	}

	finish, _ := json.Marshal(data)
	fmt.Fprintf(w, string(finish))
}

// func getSettings(w http.ResponseWriter, r *http.Request) {
// 	const ref string = "User get settings API:"
// 	user, err := UserByHeader(r)
// 	if err != nil {
// 		sendAnswerError(ref, err, "", invalidToken, 1, w)
// 		return
// 	}
// 	setts, err := db.GetUserSettings(user.ID)
// 	if err != nil {
// 		sendAnswerError(ref, err, fmt.Sprintf("userID - %d", user.ID), failedGetSettings, 2, w)
// 		return
// 	}
// 	finish, _ := json.Marshal(setts)
// 	fmt.Fprintf(w, string(finish))
// }

func setSettings(w http.ResponseWriter, r *http.Request) {
	const ref string = "User set settings API:"
	var data struct {
		Token    string `json:"token"`
		Name     string `json:"name"`
		Language string `json:"language"`
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	user, err := TestUserToken(data.Token)
	if err != nil {
		sendAnswerError(ref, err, data.Token, invalidToken, 1, w)
		return
	}
	err = db.SetUserSettigns(user.ID, models.UserSettings{Name: data.Name, Language: data.Language})
	if err != nil {
		sendAnswerError(ref, err, fmt.Sprintf("userID - %d", user.ID), failedSetUserSettings, 2, w)
		return
	}
	sendAnswerSuccess(w)
}
