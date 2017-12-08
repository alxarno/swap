package api2

import (
	"net/http"
	"encoding/json"
	"github.com/Spatium-Messenger/Server/db_api"
	"fmt"
	"github.com/AlexeyArno/Gologer"
)

func enter(w http.ResponseWriter, r *http.Request){
	var data struct{
		Login string`json:"login"`
		Pass string	`json:"pass"`
	}
	err:= json.NewDecoder(r.Body).Decode(&data);if err!=nil{
		sendAnswerError("Wrong data", 0, w)
		return
	}
	user, err:=db_api.GetUser("login", map[string]interface{}{"login": data.Login, "pass":data.Pass})
	if err!=nil{
		Gologer.PError(err.Error())
		sendAnswerError("User not found", 0, w)
		return
	}

	//if user.CheckPass(data.Pass){
	token,err:= generateToken(user.Id);if err!=nil{
		sendAnswerError("Failed token generator", 0, w)
		return
	}
	var x = make(map[string]string)
	x["token"]=token
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
	return
	//}else{
	//	sendAnswerError("Pass is invalid", 0, w)
	//	return
	//}
}

func proveToken(w http.ResponseWriter, r *http.Request){
	var userGetToken struct{
		Token string`json:"token"`
	}
	err:=getJson(&userGetToken,r); if err!=nil{
		sendAnswerError("Failed decode token",0, w)
		return
	}
	var x = make(map[string]interface{})
	_,err= TestUserToken(userGetToken.Token);if err==nil{
		x["result"]="Success"
	}else{
		x["result"]="Error"
		x["code"] = 0
	}
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func createUser(w http.ResponseWriter, r *http.Request){
	var data struct{
		Login string`json:"login"`
		Pass string`json:"pass"`
	}
	err:=getJson(&data,r);if err!=nil{
		sendAnswerError("Failed decode data",0, w)
		return
	}
	if data.Login == "" || data.Pass == ""{
		sendAnswerError("Haven't all fields (Login,Pass)",0, w)
		return
	}
	id,err:=db_api.CreateUser(data.Login, data.Pass,data.Login);if err!=nil{
		sendAnswerError("Failed create user", 0, w)
		return
	}
	token,err:= generateToken(id);if err!=nil{
		sendAnswerError("Failed token generator", 0, w)
		return
	}
	var x = make(map[string]string)
	x["token"]=token
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
	return
}

func getMyChats(w http.ResponseWriter, r *http.Request){
	user,err:= getUserByToken(r);if err!=nil{
		sendAnswerError(err.Error(),0,w); return
	}
	chats,err:= db_api.GetUserChats(user.Id);if err!=nil{
		sendAnswerError(err.Error(),0,w); return
	}
	var finish []byte
	if chats==nil{
		finish=[]byte("[]")
	}else{
		finish, _=json.Marshal(chats)
	}
	fmt.Fprintf(w, string(finish))
}

func getMyData(w http.ResponseWriter, r *http.Request){
	user,err:= getUserByToken(r);if err!=nil{
		sendAnswerError(err.Error(),0,w); return
	}
	data := make(map[string]interface{})
	data["id"] = user.Id
	finish, _:=json.Marshal(data)
	fmt.Fprintf(w, string(finish))
}

func getSettings(w http.ResponseWriter, r *http.Request){
	user,err:=getUserByToken(r);if err!=nil{
		sendAnswerError(err.Error(),0,w); return
	}
	setts,err:= db_api.GetUserSettings(user.Id); if err!=nil{
		sendAnswerError(err.Error(),0,w); return
	}
	finish, _:=json.Marshal(setts)
	fmt.Fprintf(w, string(finish))
}

func setSettings(w http.ResponseWriter, r *http.Request){
	var data struct{Token string; Name string}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data);if err!=nil{
		sendAnswerError(err.Error(),0,w); return
	}
	user,err:= TestUserToken(data.Token);if err!=nil{
		sendAnswerError(err.Error(),0,w); return
	}
	err= db_api.SetUserSettings(user.Id, data.Name);if err!=nil{
		sendAnswerError(err.Error(),0,w); return
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
	}
}