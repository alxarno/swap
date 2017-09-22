package messages

import (
	methods "github.com/AlexArno/spatium/src/api/methods"
	db_work "github.com/AlexArno/spatium/db_work"
	"encoding/json"
	"errors"
	"github.com/AlexArno/spatium/models"
)
var secret = "321312421"
func Hello() string{
	return "Hello"
}

type NewMessageFormUser struct{
	Chat_Id *float64
	Content *models.MessageContent
	Token *string
}



func NewMessage(user_quest *string)(*models.NewMessageToUser, error){
	var data *NewMessageFormUser
	message := []byte(*user_quest)
	err := json.Unmarshal(message, &data)
	if err != nil{
		return nil,err
	}
	if data.Chat_Id == nil {
		return nil, errors.New("Chat_Id is missing or null!")
	}
	if data.Token == nil {
		return nil, errors.New("Token is missing or null!")
	}
	if data.Content  == nil {
		return nil, errors.New("Content is missing or null!")
	}
	if data.Content.Message  == nil {
		return nil, errors.New("Content.Message is missing or null!")
	}
	if data.Content.Documents  == nil {
		return nil, errors.New("Content.Documents is missing or null!")
	}
	if data.Content.Type  == nil {
		return nil, errors.New("Content.Type is missing or null!")
	}
	token := *data.Token
	user,err := methods.TestUserToken(secret, token)
	if err != nil{
		return nil, err
	}
	content,err:= json.Marshal(*data.Content)
	if err!=nil{
		return  nil,err
	}
	err= db_work.AddMessage(user.ID, *data.Chat_Id, string(content))
	if err != nil{
		return nil,err
	}
	var send *models.NewMessageToUser
	send.Author_id = &user.ID
	send.Author_Name=&user.Name
	send.Chat_Id = data.Chat_Id
	send.Content = data.Content
	return send,nil

}
