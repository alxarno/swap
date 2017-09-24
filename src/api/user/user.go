package user

import (
	//"fmt"
	"net/http"
	"encoding/json"
	db_work "github.com/AlexArno/spatium/db_work"
	"github.com/robbert229/jwt"
	"time"
	methods "github.com/AlexArno/spatium/src/api/methods"
	"strconv"

	"fmt"

	"os"
	"io"
)
type ProveConnection struct{
	Login string
	Pass string
}
type CreateNewUser struct{
	Login string
	Pass string
	Name string
}
type userGetToken struct{
	Token string
}

type TokenData struct{
	Id int
	Time int
}

var (
	secret = "321312421"
)
func sendAnswerError(e_type string, w http.ResponseWriter){
	var answer = make(map[string]string)
	answer["result"] = "Error"
	answer["type"]=e_type
	finish, _:=json.Marshal(answer)
	fmt.Fprintf(w, string(finish))
}

func sendToken(id string, w http.ResponseWriter){
	algorithm :=  jwt.HmacSha256(secret)
	claims := jwt.NewClaim()
	claims.Set("id", id)
	claims.Set("time", time.Now().AddDate(0,0,30).Unix())
	token, err := algorithm.Encode(claims)
	if err!=nil{
		sendAnswerError("Token is failed", w)
		//fmt.Println(err)
		return
	}
	var x = make(map[string]string)
	x["token"]=token
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}


func getJson(target interface{}, r*http.Request) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func uploadFile(w http.ResponseWriter, r *http.Request){
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseMultipartForm(104857600)
	fmt.Println(r.FormValue("token"))
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile("./public/files/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	//if r.Method != "GET" {
	//	r.ParseMultipartForm(32 << 20)
	//	token := r.FormValue("token")
	//	fmt.Println(token)
	//	file, handler, err := r.FormFile("uploadfile")
	//	if err != nil {
	//		fmt.Println("first", err)
	//		return
	//	}
	//	defer file.Close()
	//	fmt.Fprintf(w, "%v", handler.Header)
	//	f, err := os.OpenFile("./public/files/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	//	if err != nil {
	//		fmt.Println("second", err)
	//		return
	//	}
	//	defer f.Close()
	//	io.Copy(f, file)
	//}
}


func enter( w http.ResponseWriter, r *http.Request){
	var data *ProveConnection
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	//fmt.Println(data)
	if data == nil{
		sendAnswerError("Haven't all fields (login, pass)", w)
		return
	}
	now_user, err:= db_work.GetUser("login" , map[string]string{"login":data.Login, "pass":data.Pass})
	if err!=nil{
		sendAnswerError("User is undefined", w)
		return
	}
	if now_user == nil{
		sendAnswerError("User is undefined", w)
		return
	}
	id_int64 := int64(now_user.ID)
	u_id:= strconv.FormatInt(id_int64, 10)
	sendToken(u_id, w)
}

func proveToken(w http.ResponseWriter, r *http.Request){
	var data *userGetToken
	err:=getJson(&data,r)
	if err != nil {
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	if data == nil{
		sendAnswerError("Haven't all fields (Token)", w)
		return
	}
	tokenIsTrue, err := methods.TestUserToken(secret, data.Token)
	if err != nil {
		methods.SendAnswerError(err.Error(), w)
		return
	}
	if tokenIsTrue != nil{
		var x = make(map[string]string)
		x["result"]="Success"
		finish, _:=json.Marshal(x)
		fmt.Fprintf(w, string(finish))
	}
}

func createUser(w http.ResponseWriter, r *http.Request){
	var data *ProveConnection
	err:=getJson(&data,r)
	if err != nil {
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	if data.Login == "" || data.Pass == ""{
		sendAnswerError("Haven't all fields (Login,Pass)", w)
		return
	}
	id,err_string, err := db_work.CreateUser(data.Login, data.Pass, data.Login)
	if err != nil || id==""{
			sendAnswerError(err_string,w)
			return
	}
	sendToken(id, w)
}

func getMyChats(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	user, err:=methods.DecodeToken(secret, r)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	chats,err := db_work.GetMyChats(user.ID)
	if err!= nil{
		sendAnswerError("Some failed",w)
		return
	}
	finish, _:=json.Marshal(chats)
	fmt.Fprintf(w, string(finish))
}

func getMyData(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	user, err:=methods.DecodeToken(secret, r)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	if user == nil{
		sendAnswerError("User is undefined", w)
		return
	}

	data := make(map[string]interface{})
	data["ID"] = user.ID
	finish, _:=json.Marshal(data)
	fmt.Fprintf(w, string(finish))
}

func MainUserApi(var1 string, w http.ResponseWriter, r *http.Request){
	//fmt.Println(var1+"Hello")
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
	case "uploadFile":
		uploadFile(w,r)
	}
}
