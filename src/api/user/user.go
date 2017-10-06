package user

import (
	//"fmt"
	"net/http"
	"encoding/json"
	db_work "github.com/AlexeyArno/Spatium/db_work"
	engine "github.com/AlexeyArno/Spatium/src/message_engine"
	"github.com/robbert229/jwt"
	"time"
	methods "github.com/AlexeyArno/Spatium/src/api/methods"
	"strconv"
	"fmt"
	"os"
	"io"
	"github.com/AlexeyArno/Spatium/models"
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

func getDisposableFileLink(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data *struct{Token string; FileId string}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.OnlyDecodeToken(secret, data.Token)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	path, err := db_work.GetFileProve(user.ID, data.FileId)
	if err!=nil{
		sendAnswerError(err.Error(), w)
		return
	}
	algorithm :=  jwt.HmacSha256(secret)
	claims := jwt.NewClaim()
	claims.Set("path", path)
	claims.Set("user_id", user.ID)
	claims.Set("time", time.Now().Unix()+60)
	link, err := algorithm.Encode(claims)
	if err!=nil{
		fmt.Fprintf(w, "Faile encode link")
		//fmt.Println(err)
		return
	}
	var x = make(map[string]string)
	x["link"]=link
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func uploadFile(w http.ResponseWriter, r *http.Request){
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	r.ParseMultipartForm(104857600)
	s_ratio_size := r.FormValue("ratio_size")
	fmt.Println(s_ratio_size)
	ratio_size,err := strconv.ParseFloat(s_ratio_size,64)
	if err !=nil{
		sendAnswerError(err.Error(), w)
		return
	}
	//fmt.Println(i_type)
	token := r.FormValue("token")
	i_type := r.FormValue("type")
	chat_id :=  r.FormValue("chat_id")
	user, err:=methods.OnlyDecodeToken(secret, token)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		sendAnswerError(err.Error(), w)
		fmt.Println(err)
		return
	}

	defer file.Close()

	id, path, err := db_work.CreateFile(handler.Filename, handler.Size,user.ID, chat_id, s_ratio_size)
	if err != nil{
		sendAnswerError(err.Error(), w)
		fmt.Println(err)
		return
	}
	f, err := os.OpenFile("./public/files/"+path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		sendAnswerError(err.Error(), w)
		fmt.Println(err)
		return
	}

	defer f.Close()
	io.Copy(f, file)
	go MiniMize(i_type,ratio_size,path)
	var x = make(map[string]string)
	x["result"]="Success"
	x["FileId"]= strconv.FormatInt(id,10)
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))


}

func deleteFile(w http.ResponseWriter, r *http.Request){
	var data *struct{Token string; FileId string}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.OnlyDecodeToken(secret, data.Token)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	path, err := db_work.DeleteFile(user.ID, data.FileId)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	//fmt.Println(path)
	if path == ""{
		sendAnswerError("Path is undefined", w)
		return
	}
	err = os.Remove("./public/files/"+path)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	err = os.Remove("./public/files/min/"+path)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	var x = make(map[string]string)
	x["result"]="Success"
	finish, _:=json.Marshal(x)
	fmt.Fprintf(w, string(finish))
}

func getFile(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data *struct{Token string; FileId string; Min int}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.OnlyDecodeToken(secret, data.Token)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	path, err := db_work.GetFileProve(user.ID, data.FileId)
	if err!=nil{
		sendAnswerError(err.Error(), w)
		return
	}
	file := "./public/files/"+path
	if data.Min == 1{
		file =	"./public/files/min/"+path
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		// path/to/whatever does not exist
		if data.Min ==1{
			file = "./public/files/"+path
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return
			}
		}
	}

	http.ServeFile(w, r, file)
	//w.started = true
	return
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
	for _,v:= range chats{
		v.Online, _ = getOnlineUsersIntChat(v.ID)
	}
	finish, _:=json.Marshal(chats)
	fmt.Fprintf(w, string(finish))
}

func getOnlineUsersIntChat(chat_id float64)(int64, error){
	var users_id []float64
	users_id,_ = db_work.GetChatsUsers(chat_id)
	final:= engine.GetOnlineUsersInChat(&users_id)
	return final, nil
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

func getUsersForAdd(w http.ResponseWriter, r *http.Request){
	var data *struct{Token string; Chatid string; Name string}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	users ,err:= db_work.FindUserByName(data.Name, data.Chatid)
	if err!=nil{
		fmt.Println(err)
		sendAnswerError("Failed find user", w)
		return
	}
	f_data := make(map[string]interface{})
	f_data["result"] = "Success"
	f_data["users"] = users
	finish, _:=json.Marshal(f_data)
	fmt.Fprintf(w, string(finish))

}

func getSettings(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data *struct{Token string}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.OnlyDecodeToken(secret, data.Token)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	final,err := db_work.GetUserSettings(user.ID)
	if err!=nil{
		sendAnswerError(err.Error(), w)
		return
	}
	finish, _:=json.Marshal(final)
	fmt.Fprintf(w, string(finish))
}

func SetSettings(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data *struct{Token string; Name string}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.OnlyDecodeToken(secret, data.Token)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	err = db_work.SetUserSettings(user.ID, data.Name)
	if err!=nil{
		sendAnswerError(err.Error(), w)
		return
	}
	var answer = make(map[string]string)
	answer["result"] = "Success"
	finish, _:=json.Marshal(answer)
	fmt.Fprintf(w, string(finish))
}
//
func GetUsersByName(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data *struct{Token string; Name string}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.OnlyDecodeToken(secret, data.Token)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	list,err:= db_work.GetUsersForCreateDialog(user.ID,data.Name)
	if err!= nil{
		sendAnswerError(err.Error(), w)
		return
	}
	finish, _:=json.Marshal(list)
	fmt.Fprintf(w, string(finish))
}

func CreateDialog(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data *struct{Token string; User_id float64}
	decoder:= json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println(err)
		sendAnswerError("Failed decode r.Body", w)
		return
	}
	user, err:=methods.OnlyDecodeToken(secret, data.Token)
	if err != nil{
		sendAnswerError(err.Error(), w)
		return
	}
	msg, ch_id,int_last,err:= db_work.CreateDialog(user.ID,data.User_id)
	if err!=nil{
		sendAnswerError(err.Error(), w)
		return
	}
	content:= models.MessageContentToUser{msg.Message, []interface{}{},msg.Type}
	now:= time.Now().Unix()
	send_message:= models.NewMessageToUser{&int_last,&ch_id, content,&user.ID,&user.Name,&now}
	engine.SendMessage(send_message)
	engine.SendNotificationAddUserInChat(data.User_id)
	var answer = make(map[string]string)
	answer["result"] = "Success"
	finish, _:=json.Marshal(answer)
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
	case "deleteFile":
		deleteFile(w, r)
	case "getFile":
		getFile(w,r)
	case "getFileLink":
		getDisposableFileLink(w,r)
	case "getUsersForAdd":
		getUsersForAdd(w,r)
	case "getSettings":
		getSettings(w,r)
	case "setSettings":
		SetSettings(w,r)
	case "getUsersByNameForCreateDialog":
		GetUsersByName(w,r)
	case "createDialog":
		CreateDialog(w,r)
	}
}
