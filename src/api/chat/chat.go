package chat

import (
	"net/http"
	methods "github.com/AlexArno/spatium/src/api/methods"
	db_work "github.com/AlexArno/spatium/db_work"
	"strconv"
	"encoding/json"
	"fmt"
)
var secret = "321312421"

type createChat struct{
	Token string
	//Login string
	Name string
}

type getMessagesData struct{
	Token string
	ID float64
}

func create(w http.ResponseWriter, r *http.Request){
	var data *createChat
	err:=methods.GetJson(&data, r)
	if err != nil {
		methods.SendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.TestUserToken(secret, data.Token)
	if err != nil{
		methods.SendAnswerError("Failed decode token", w)
		return
	}
	id_int64 := int64(user.ID)
	u_id:= strconv.FormatInt(id_int64, 10)
	_,err = db_work.CreateChat(data.Name, u_id)
	if err != nil{
		methods.SendAnswerError(err.Error(), w)
		return
	}
	var x = make(map[string]string)
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func getMessages(w http.ResponseWriter, r *http.Request){
	var data *getMessagesData
	err:=methods.GetJson(&data, r)
	//fmt.Println(data)
	if err != nil {
		methods.SendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.TestUserToken(secret, data.Token)
	if err != nil{
		methods.SendAnswerError("Failed decode token", w)
		return
	}
	err = db_work.CheckUserINChat(user.ID, data.ID)
	if err != nil{
		methods.SendAnswerError("User isn't in chat", w)
		return
	}
	messages,err:=db_work.GetMessages(data.ID)
	if err != nil{
		fmt.Println(err.Error())
		methods.SendAnswerError("Fail get data from db", w)
		return
	}
	if messages == nil{
		finish, _:=json.Marshal([]string{})
		fmt.Fprintf(w, string(finish))
		return
	}
	finish, _:=json.Marshal(messages)
	fmt.Fprintf(w, string(finish))
}

func MainChatApi(var1 string, w http.ResponseWriter, r *http.Request){
	switch var1 {
	case "create":
		create(w,r)
	case "getMessages":
		getMessages(w,r)
	}
}
