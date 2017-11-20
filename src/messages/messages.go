package messages

import (
	methods "github.com/Spatium-Messenger/Server/src/api/methods"
	db_work "github.com/Spatium-Messenger/Server/db_work"
	"encoding/json"
	"errors"
	"github.com/Spatium-Messenger/Server/models"
	"github.com/Spatium-Messenger/Server/settings"
	"fmt"
)
var secret = settings.ServiceSettings.Server.SecretKeyForToken


type NewMessageFormUser struct{
	Chat_Id *float64
	Content *models.MessageContent
	Token *string
}



func NewMessage(user_quest *string)(models.NewMessageToUser, error){
	var send models.NewMessageToUser
	var data NewMessageFormUser
	message := []byte(*user_quest)
	err := json.Unmarshal(message, &data)
	if err != nil{
		return send,err
	}
	if data.Chat_Id == nil {
		return send, errors.New("Chat_Id is missing or null!")
	}
	if data.Token == nil {
		return send, errors.New("Token is missing or null!")
	}
	if data.Content  == nil {
		return send, errors.New("Content is missing or null!")
	}
	if data.Content.Message  == nil {
		return send, errors.New("Content.Message is missing or null!")
	}
	if data.Content.Documents  == nil {
		return send, errors.New("Content.Documents is missing or null!")
	}
	if data.Content.Type  == nil {
		return send, errors.New("Content.Type is missing or null!")
	}
	token := *data.Token
	user,err := methods.TestUserToken(secret, token)
	if err != nil{
		return send, err
	}
	content,err:= json.Marshal(*data.Content)
	if err!=nil{
		return  send,err
	}
	m_id,err:= db_work.AddMessage(user.ID, *data.Chat_Id, string(content))
	if err != nil{
		return send,err
	}
	newContent,err := methods.ProcessMessageFromUserToUser( data.Content)
	if err != nil{
		return  send,err
	}
	//fmt.Println(newContent)
	send.ID = &m_id
	send.Author_id = &user.ID
	send.Author_Name=&user.Name
	send.Chat_Id = data.Chat_Id
	send.Content = newContent
	return send, nil


}

func NewMessageAnotherStruct(user_quest *string)(models.NewMessageToUser, error){
	var send models.NewMessageToUser
	var data struct{
		Type string
		Content NewMessageFormUser
	}
	message := []byte(*user_quest)
	err := json.Unmarshal(message, &data)
	if err != nil{
		return send,err
	}
	if data.Content.Chat_Id == nil {
		return send, errors.New("Chat_Id is missing or null!")
	}
	if data.Content.Token == nil {
		return send, errors.New("Token is missing or null!")
	}
	if data.Content.Content  == nil {
		return send, errors.New("Content is missing or null!")
	}
	if data.Content.Content.Message  == nil {
		return send, errors.New("Content.Message is missing or null!")
	}
	if data.Content.Content.Documents  == nil {
		return send, errors.New("Content.Documents is missing or null!")
	}
	if data.Content.Content.Type  == nil {
		return send, errors.New("Content.Type is missing or null!")
	}
	token := *data.Content.Token
	user,err := methods.TestUserToken(secret, token)
	if err != nil{
		return send, err
	}
	content,err:= json.Marshal(*data.Content.Content)
	if err!=nil{
		fmt.Println("113")
		return  send,err
	}
	m_id,err:= db_work.AddMessage(user.ID, *data.Content.Chat_Id, string(content))
	//f_m_id:= float64(m_id)
	if err != nil{
		fmt.Println(err.Error())
		return send,err
	}
	newContent,err := methods.ProcessMessageFromUserToUser( data.Content.Content)
	if err != nil{
		fmt.Println(err.Error())
		return  send,err
	}
	//fmt.Println(newContent)
	send.ID = &m_id
	send.Author_id = &user.ID
	send.Author_Name=&user.Name
	send.Author_Login = &user.Login
	send.Chat_Id = data.Content.Chat_Id
	send.Content = newContent
	return send, nil
}

//func NewMessagev2(msg *string)
