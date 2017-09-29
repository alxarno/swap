package chat

import (
	"net/http"
	methods "github.com/AlexArno/spatium/src/api/methods"
	db_work "github.com/AlexArno/spatium/db_work"
	engine "github.com/AlexArno/spatium/src/message_engine"
	"strconv"
	"encoding/json"
	"fmt"
	"github.com/AlexArno/spatium/models"
	"time"
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
	messages,err:=db_work.GetMessages(user.ID,data.ID)
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

func addUsers(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data struct{Token string; Ids []string; ChatId string}
	err:=methods.GetJson(&data, r)
	if err != nil {
		methods.SendAnswerError("Failed decode r.Body", w)
		return
	}
	//fmt.Println(data)
	user,err := methods.OnlyDecodeToken(secret, data.Token)
	if err != nil {
		fmt.Println(err)
		methods.SendAnswerError("Failed decode token", w)
		return
	}
	// Because we need it
	f64_caht_id,err := strconv.ParseFloat(data.ChatId, 64)
	if err != nil{
		fmt.Println("FAIL DECODE Ids")
		return
	}

	err = db_work.CheckUserINChat(user.ID, f64_caht_id)
	if err != nil{
		return
	}


	i_chat_add, err := strconv.ParseInt(data.ChatId, 10, 64)
	if err != nil {
		fmt.Println(err)
		methods.SendAnswerError("Failed retype chat_id", w)
		return
	}
	failesAdd := []string{}
	successAdd := []string{}
	for i:=0;i<len(data.Ids);i++{
		err:= db_work.InsertUserInChat(data.Ids[i], i_chat_add)
		if err!=nil{
			fmt.Println(err)
			//methods.SendAnswerError("Fail add in chat "+data.Ids[i], w)
			failesAdd= append(failesAdd, data.Ids[i])
		}else{
			successAdd = append(successAdd, data.Ids[i])
		}
	}
	//If count success added users not equal all need add user we send error to user
	if len(successAdd) != len(data.Ids){
		//return
		var final = make(map[string]interface{})
		final["result"] = "Error"
		final["failed"] = failesAdd
		//final[""]
		finish, _:=json.Marshal(final)
		fmt.Fprintf(w, string(finish))
		return
	}
	var final = make(map[string]interface{})
	final["result"] = "Success"
	final["success"] = successAdd
	//final[""]
	finish, _:=json.Marshal(final)
	fmt.Fprintf(w, string(finish))
	//Send notification and messages to users and chats
	for _,v := range data.Ids{
		id,err := strconv.ParseFloat(v, 64)
		if err != nil{
			fmt.Println("FAIL DECODE Ids")
		}
		add_user_info,err := db_work.GetUser("id", map[string]string{"id": v})
		if err != nil{
			fmt.Println("FAIL GET data by id")
		}
		docs:= make([]interface{},0)
		msg_content:= "пригласил "+add_user_info.Name
		str:= "a_msg"
		message:= models.MessageContentToUser{&msg_content, docs, &str}
		s_message,err :=json.Marshal(message)
		if err != nil{
			fmt.Println("FAIL MARSHAL MessageContentToUser")
		}
		now_time:=time.Now().Unix()
		db_work.AddMessage(user.ID,f64_caht_id,string(s_message))
		send_message:= models.NewMessageToUser{&f64_caht_id, message,&user.ID,&user.Name,&now_time}
		engine.SendMessage(send_message)
		engine.SendNotificationAddUserInChat(id)
	}

}

func getUsers(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data struct{Token string; ChatId string}
	err:=methods.GetJson(&data, r)
	if err != nil {
		methods.SendAnswerError("Failed decode r.Body", w)
		return
	}
	//fmt.Println(data)
	user,err := methods.OnlyDecodeToken(secret, data.Token)
	if err != nil {
		fmt.Println(err)
		methods.SendAnswerError("Failed decode token", w)
		return
	}
	f64_caht_id,err := strconv.ParseFloat(data.ChatId, 64)
	if err != nil{
		fmt.Println("FAIL DECODE CHAT ID")
		return
	}
	err= db_work.CheckUserINChat(user.ID, f64_caht_id)
	if err!=nil{
		fmt.Println("FAIL GetChatsUsers")
		methods.SendAnswerError("You aren't have rights for this action", w)
		return
	}

	users,err:=db_work.GetChatUsersInfo(f64_caht_id)
	if err != nil{
		fmt.Println("FAIL GetChatsUsers")
		return
	}
	//finish, _:=json.Marshal(users)
	fmt.Fprintf(w, string(users))
	return
}

func deleteUsers(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data struct{Token string; Ids []float64; ChatId string}
	err:=methods.GetJson(&data, r)
	if err != nil {
		methods.SendAnswerError("Failed decode r.Body", w)
		return
	}
	//fmt.Println(data)
	user,err := methods.OnlyDecodeToken(secret, data.Token)
	if err != nil {
		fmt.Println(err)
		methods.SendAnswerError("Failed decode token", w)
		return
	}

	// Because we need it
	f64_caht_id,err := strconv.ParseFloat(data.ChatId, 64)
	if err != nil{
		fmt.Println("FAIL DECODE Ids")
		return
	}
	err= db_work.CheckUserRightsInChat(user.ID, f64_caht_id)
	if err !=nil{
		methods.SendAnswerError(err.Error(), w)
		return
	}

	err = db_work.CheckUserRightsInChat(user.ID, f64_caht_id)
	if err!=nil{
		methods.SendAnswerError(err.Error(), w)
		return
	}

	err = db_work.DeleteUsersInChat(data.Ids, data.ChatId)
	if err!=nil{
		methods.SendAnswerError(err.Error(), w)
		return
	}
	for _,v:=range data.Ids{
		s_id := strconv.FormatFloat(v,'f',0,64)
		add_user_info,err := db_work.GetUser("id", map[string]string{"id": s_id})
		if err != nil{
			fmt.Println("FAIL GET data by id")
		}
		docs:= make([]interface{},0)
		msg_content:= "удалил "+add_user_info.Name
		str:= "a_msg"
		message:= models.MessageContentToUser{&msg_content, docs, &str}
		s_message,err :=json.Marshal(message)
		if err != nil{
			fmt.Println("FAIL MARSHAL MessageContentToUser")
		}
		now_time:=time.Now().Unix()
		db_work.AddMessage(user.ID,f64_caht_id,string(s_message))
		send_message:=models.NewMessageToUser{&f64_caht_id, message,&user.ID,&user.Name,&now_time}
		force_msg:=models.ForceMsgToUser{v,send_message}
		engine.SendForceMessage(force_msg)
		engine.SendMessage(send_message)
		engine.SendNotificationAddUserInChat(v)
	}
	finish:= make(map[string]string)
	finish["result"] = "Success"
	final, err := json.Marshal(finish)
	fmt.Fprintf(w, string(final))



}


func MainChatApi(var1 string, w http.ResponseWriter, r *http.Request){
	switch var1 {
	case "create":
		create(w,r)
	case "getMessages":
		getMessages(w,r)
	case "addUsersInChat":
		addUsers(w,r)
	case "getUsers":
		getUsers(w,r)
	case "deleteUsers":
		deleteUsers(w,r)
	}
}
