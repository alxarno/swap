package message_engine

import (
	"encoding/json"
	"fmt"
	"github.com/AlexArno/spatium/models"
	"github.com/AlexArno/spatium/src/api/methods"
	messages_work "github.com/AlexArno/spatium/src/messages"
	"time"
)
var (
	secret = "321312421"
)
func SystemMsg(msg string)(map[string]interface{}, error){
	var final = make(map[string]interface{})
	var user_msg_sys = struct {
		Type string
		Content struct{Type string; Token string}
	}{}
	if err := json.Unmarshal([]byte(msg), &user_msg_sys); err != nil {
		//panic(err)
		fmt.Println(err)
		return nil, err
	}
	if user_msg_sys.Content.Type == "authoriz"{
		_,err:= methods.TestUserToken(secret, user_msg_sys.Content.Token)
		if err!= nil{
			fmt.Println(err)
			return nil, err
		}
		final["Action"] = "Authoriz"
		final["Payload"] = user_msg_sys.Content.Token
		return final,nil

	}
	return nil,nil
}

func UserMsg(msg string)(*models.NewMessageToUser, error){
	//var user_msg = struct {
	//	Type string
	//	Content models
	//}{}

	message,err:= messages_work.NewMessageAnotherStruct(&msg)
	if err !=nil{
		return nil,err
	}
	now_time :=  time.Now().Unix()
	message.Time =  &now_time
	return &message,nil
	//if err := json.Unmarshal([]byte(msg), &user_msg); err != nil {
	//	//panic(err)
	//	fmt.Println(err)
	//	return nil,err
	//}
	////{"Chat_Id":2,"Content":{"Message":"...","Documents":["1","2"],"Type":"u_msg"},"Token":"eyJUeXA..."}
	//user_s_msg := make(map[string]interface{})
	//
	//user_s_msg["Chat_Id"] = user_msg.Content.Chat_Id
	//user_s_msg["Content"] = user_msg.Content.Content
	//user_s_msg["Token"] = user_msg.Content.Token
	//jsonMessageContent, err:= json.Marshal(user_s_msg)
	//if err != nil{
	//	return nil,err
	//}
	//s_js_msg := jsonMessageContent
	//newMsgToUser,err:=messages_work.NewMessage([]byte(s_js_msg))


}