package messages

import (
	"encoding/json"

	// "github.com/AlexeyArno/Gologer"
	"github.com/swap-messenger/Backend/db"
	"github.com/swap-messenger/Backend/models"
	"github.com/swap-messenger/Backend/src/api"
)

type NewMessageFormUser struct {
	ChatId  int64                  `json:"Chat_Id"`
	Content *models.MessageContent `json:"Content"`
	Token   string                 `json:"Token"`
}

type NewMessageReceive struct {
	ChatId  int64                  `json:"Chat_Id"`
	Content *MessageContentReceive `json:"Content"`
	Token   string                 `json:"Token"`
}

type MessageContentReceive struct {
	Message   string    `json:"content"`
	Documents []float64 `json:"documents"`
	Type      string    `json:"type"`
}

func NewMessage(userQuest *string) (models.NewMessageToUser, error) {
	var send models.NewMessageToUser
	var data NewMessageFormUser
	message := []byte(*userQuest)
	err := json.Unmarshal(message, &data)
	if err != nil {
		// Gologer.PError(err.Error())
		return send, err
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
	user, err := api.TestUserToken(data.Token)
	if err != nil {
		// Gologer.PError(err.Error())
		return send, err
	}
	content, err := json.Marshal(*data.Content)
	if err != nil {
		// Gologer.PError(err.Error())
		return send, err
	}
	messageId, err := db.SendMessage(user.Id, data.ChatId, string(content), 0, 0)
	if err != nil {
		// Gologer.PError(err.Error())
		return send, err
	}
	//newContent,err := methods.ProcessMessageFromUserToUser( data.Content)
	//if err != nil{
	//	return  send,err
	//}
	//fmt.Println(newContent)

	//Get file information
	var documents []map[string]interface{}

	for _, v := range data.Content.Documents {
		doc, err := db.GetFileInformation(v)
		if err != nil {
			continue
		}
		documents = append(documents, doc)
	}

	var newMess models.MessageContentToUser

	newMess.Message = data.Content.Message
	newMess.Type = data.Content.Type
	newMess.Documents = documents

	send.ID = messageId
	send.AuthorId = user.Id
	send.AuthorName = user.Name
	send.ChatId = data.ChatId
	send.Content = &newMess
	return send, nil

}

func NewMessageAnother(userQuest *string) (models.NewMessageToUser, error) {
	// log.Println(*userQuest)
	var send models.NewMessageToUser
	var dataReceive struct {
		Type    string
		Content NewMessageReceive
	}

	message := []byte(*userQuest)
	err := json.Unmarshal(message, &dataReceive)
	if err != nil {
		return send, err
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
	user, err := api.TestUserToken(dataReceive.Content.Token)
	if err != nil {
		// Gologer.PError(err.Error())
		return send, err
	}

	// Gologer.PInfo(strconv.FormatInt(user.Id, 10))
	//content,err:= json.Marshal(*data.Content.Content);if err!=nil{
	//	//Gologer.PError(err.Error())
	//	return  send,err
	//}
	messageCon, err := json.Marshal(dataReceive.Content.Content)
	if err != nil {
		// Gologer.PError(err.Error())
		return send, err
	}
	mId, err := db.SendClearMessage(user.Id, dataReceive.Content.ChatId, string(messageCon))
	if err != nil {
		//Gologer.PError(err.Error())
		return send, err
	}
	//newContent,err := methods.ProcessMessageFromUserToUser( data.Content.Content)
	//if err != nil{
	//	fmt.Println(err.Error())
	//	return  send,err
	//}
	//fmt.Println(newContent)
	var documents []map[string]interface{}

	for _, v := range dataReceive.Content.Content.Documents {
		doc, err := db.GetFileInformation(int64(v))
		if err != nil {
			continue
		}
		documents = append(documents, doc)
	}

	var newMess models.MessageContentToUser

	newMess.Message = dataReceive.Content.Content.Message
	newMess.Type = dataReceive.Content.Content.Type
	newMess.Documents = documents

	send.ID = mId
	send.AuthorId = user.Id
	send.AuthorName = user.Name
	send.ChatId = dataReceive.Content.ChatId
	send.Content = &newMess
	return send, nil

}

//func NewMessagev2(msg *string)
