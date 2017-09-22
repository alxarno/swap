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

func MainChatApi(var1 string, w http.ResponseWriter, r *http.Request){
	switch var1 {
	case "create":
		create(w,r)

	}
}
