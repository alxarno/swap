package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/swap-messenger/swap/models"

	db "github.com/swap-messenger/swap/db2"
)

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
	mapData := map[string]interface{}{"login": data.Login, "pass": data.Pass}
	user, err := db.GetUserByLoginAndPass(data.Login, data.Pass)
	if err != nil {
		sendAnswerError(ref, err, mapData, FAILED_GET_USER, 1, w)
		return
	}

	//if user.CheckPass(data.Pass){
	token, err := generateToken(user.ID)
	if err != nil {
		sendAnswerError(ref, err, user.ID, FAILED_GENERATE_TOKEN, 2, w)
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
	var userGetToken struct {
		Token string `json:"token"`
	}
	err := getJson(&userGetToken, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	var x = make(map[string]interface{})
	_, err = TestUserToken(userGetToken.Token)
	if err == nil {
		x["result"] = successResult
	} else {
		x["result"] = errorResult
		x["code"] = 0
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
	err := getJson(&data, r)
	if err != nil {
		decodeFail(ref, err, r, w)
		return
	}
	if data.Login == "" || data.Pass == "" {
		sendAnswerError(ref, err, map[string]interface{}{"login": data.Login, "pass": data.Pass}, SOME_EMPTY_FIELDS, 1, w)
		return
	}
	id, err := db.CreateUser(data.Login, data.Pass, data.Login)
	if err != nil {
		sendAnswerError(ref, err, map[string]interface{}{"login": data.Login, "pass": data.Pass}, FAILED_CREATE_USER, 2, w)
		return
	}
	token, err := generateToken(id)
	if err != nil {
		sendAnswerError(ref, err, id, FAILED_GENERATE_TOKEN, 3, w)
		return
	}
	var x = make(map[string]string)
	x["token"] = token
	x["result"] = successResult
	finish, _ := json.Marshal(x)
	fmt.Fprintf(w, string(finish))
	return
}

func getMyChats(w http.ResponseWriter, r *http.Request) {
	const ref string = "User get chats API:"
	user, err := getUserByToken(r)
	if err != nil {
		sendAnswerError(ref, err, nil, INVALID_TOKEN, 1, w)
		return
	}
	chats, err := db.GetUserChats(user.ID)
	if err != nil {
		sendAnswerError(ref, err, user.ID, FAILED_GET_USER_CHATS, 2, w)
		return
	}
	var finish []byte
	if chats == nil {
		log.Println("empty chats")
		finish = []byte("[]")
	} else {
		log.Println("Not empty chats")
		finish, _ = json.Marshal(*chats)
	}
	fmt.Fprintf(w, string(finish))
}

func getMyData(w http.ResponseWriter, r *http.Request) {
	const ref string = "User get data API:"
	user, err := getUserByToken(r)
	if err != nil {
		sendAnswerError(ref, err, nil, INVALID_TOKEN, 1, w)
		return
	}
	data := make(map[string]interface{})
	data["id"] = user.ID
	finish, _ := json.Marshal(data)
	fmt.Fprintf(w, string(finish))
}

func getSettings(w http.ResponseWriter, r *http.Request) {
	const ref string = "User get settings API:"
	user, err := getUserByToken(r)
	if err != nil {
		sendAnswerError(ref, err, nil, INVALID_TOKEN, 1, w)
		return
	}
	setts, err := db.GetUserSettings(user.ID)
	if err != nil {
		sendAnswerError(ref, err, nil, FAILED_GET_SETTINGS, 2, w)
		return
	}
	finish, _ := json.Marshal(setts)
	fmt.Fprintf(w, string(finish))
}

func setSettings(w http.ResponseWriter, r *http.Request) {
	const ref string = "User set settings API:"
	var data struct {
		Token string
		Name  string
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
		sendAnswerError(ref, err, data.Token, INVALID_TOKEN, 1, w)
		return
	}
	err = db.SetUserSettigns(user.ID, models.UserSettings{Name: data.Name})
	if err != nil {
		sendAnswerError(ref, err, nil, FAILED_SET_SETTINGS, 2, w)
		return
	}
	sendAnswerSuccess(w)
}

func UserApi(var1 string, w http.ResponseWriter, r *http.Request) {
	switch var1 {
	case "enter":
		enter(w, r)
	case "testToken":
		proveToken(w, r)
	case "create":
		createUser(w, r)
	case "getMyChats":
		getMyChats(w, r)
	case "myData":
		getMyData(w, r)
	case "getSettings":
		getSettings(w, r)
	case "setSettings":
		setSettings(w, r)
	default:
		sendAnswerError("User API Router", nil, nil, END_POINT_NOT_FOUND, 0, w)
	}
}
