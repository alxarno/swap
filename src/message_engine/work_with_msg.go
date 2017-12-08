package message_engine

import (
	"encoding/json"
	"fmt"
	"github.com/Spatium-Messenger/Server/models"
	"github.com/Spatium-Messenger/Server/src/api2"
	messagesWork "github.com/Spatium-Messenger/Server/src/messages"
	"time"
	//"github.com/Spatium-Messenger/Server/settings"
)
//var (
//	secret = settings.ServiceSettings.Server.SecretKeyForToken
//)
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
		_,err:= api2.TestUserToken(user_msg_sys.Content.Token)
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

	message,err:= messagesWork.NewMessageAnotherStruct(&msg)
	if err !=nil{
		return nil,err
	}
	message.Time =  time.Now().Unix()
	return &message,nil
	//if err := json.Unmarshal([]byte(msg), &user_msg); err != nil {
	//	//panic(err)
	//	fmt.Println(err)
	//	return nil,err
	//}
	////{"chatId":2,"Content":{"Message":"...","Documents":["1","2"],"Type":"u_msg"},"Token":"eyJUeXA..."}
	//user_s_msg := make(map[string]interface{})
	//
	//user_s_msg["chatId"] = user_msg.Content.chatId
	//user_s_msg["Content"] = user_msg.Content.Content
	//user_s_msg["Token"] = user_msg.Content.Token
	//jsonMessageContent, err:= json.Marshal(user_s_msg)
	//if err != nil{
	//	return nil,err
	//}
	//s_js_msg := jsonMessageContent
	//newMsgToUser,err:=messagesWork.NewMessage([]byte(s_js_msg))


}

