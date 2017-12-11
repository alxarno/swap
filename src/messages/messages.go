package messages

import (
	"encoding/json"
	"github.com/Spatium-Messenger/Server/models"
	"fmt"
	"github.com/Spatium-Messenger/Server/src/api2"
	"github.com/Spatium-Messenger/Server/db_api"
)



type NewMessageFormUser struct{
	chatId  int64
	Content *models.MessageContent
	Token   string
}



func NewMessage(userQuest *string)(models.NewMessageToUser, error){
	var send models.NewMessageToUser
	var data NewMessageFormUser
	message := []byte(*userQuest)
	err := json.Unmarshal(message, &data)
	if err != nil{
		return send,err
	}
	//if data.chatId == nil {
	//	return send, errors.New("chatId is missing or null!")
	//}
	//if data.Token == nil {
	//	return send, errors.New("Token is missing or null!")
	//}
	//if data.Content  == nil {
	//	return send, errors.New("Content is missing or null!")
	//}
	//if data.Content.Message  == nil {
	//	return send, errors.New("Content.Message is missing or null!")
	//}
	//if data.Content.Documents  == nil {
	//	return send, errors.New("Content.Documents is missing or null!")
	//}
	//if data.Content.Type  == nil {
	//	return send, errors.New("Content.Type is missing or null!")
	//}
	//token := *data.Token
	user,err := api2.TestUserToken(data.Token)
	if err != nil{
		return send, err
	}
	content,err:= json.Marshal(*data.Content)
	if err!=nil{
		return  send,err
	}
	messageId,err:= db_api.SendMessage(user.Id, data.chatId, string(content), 0)
	if err != nil{
		return send,err
	}
	//newContent,err := methods.ProcessMessageFromUserToUser( data.Content)
	//if err != nil{
	//	return  send,err
	//}
	//fmt.Println(newContent)
	var newMess models.MessageContentToUser

	newMess.Message = *data.Content.Message
	newMess.Type = *data.Content.Type
	newMess.Documents = *data.Content.Documents


	send.ID = messageId
	send.AuthorId = user.Id
	send.AuthorName=user.Name
	send.ChatId = data.chatId
	send.Content = &newMess
	return send, nil


}

func NewMessageAnotherStruct(userQuest *string)(models.NewMessageToUser, error){
	var send models.NewMessageToUser
	var data struct{
		Type string
		Content NewMessageFormUser
	}
	message := []byte(*userQuest)
	err := json.Unmarshal(message, &data)
	if err != nil{
		return send,err
	}
	//if data.Content.chatId == nil {
	//	return send, errors.New("chatId is missing or null!")
	//}
	//if data.Content.Token == nil {
	//	return send, errors.New("Token is missing or null!")
	//}
	//if data.Content.Content  == nil {
	//	return send, errors.New("Content is missing or null!")
	//}
	//if data.Content.Content.Message  == nil {
	//	return send, errors.New("Content.Message is missing or null!")
	//}
	//if data.Content.Content.Documents  == nil {
	//	return send, errors.New("Content.Documents is missing or null!")
	//}
	//if data.Content.Content.Type  == nil {
	//	return send, errors.New("Content.Type is missing or null!")
	//}
	//token := *data.Content.Token
	user,err := api2.TestUserToken(data.Content.Token)
	if err != nil{
		return send, err
	}
	content,err:= json.Marshal(*data.Content.Content)
	if err!=nil{
		fmt.Println("113")
		return  send,err
	}
	mId,err:= db_api.SendMessage(user.Id, data.Content.chatId, string(content),0)
	//f_m_id:= float64(mId)
	if err != nil{
		fmt.Println(err.Error())
		return send,err
	}
	//newContent,err := methods.ProcessMessageFromUserToUser( data.Content.Content)
	//if err != nil{
	//	fmt.Println(err.Error())
	//	return  send,err
	//}
	//fmt.Println(newContent)

	var newMess models.MessageContentToUser

	newMess.Message = *data.Content.Content.Message
	newMess.Type = *data.Content.Content.Type
	newMess.Documents = *data.Content.Content.Documents


	send.ID = mId
	send.AuthorId = user.Id
	send.AuthorName=user.Name
	send.ChatId = data.Content.chatId
	send.Content = &newMess
	return send, nil

}

//func NewMessagev2(msg *string)
