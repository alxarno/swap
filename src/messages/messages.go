package messages

import (
	"encoding/json"

	db "github.com/swap-messenger/swap/db2"
	"github.com/swap-messenger/swap/models"
	"github.com/swap-messenger/swap/src/api"
)

type NewMessageFormUser struct {
	ChatID  int64                  `json:"Chat_Id"`
	Content *models.MessageContent `json:"Content"`
	Token   string                 `json:"Token"`
}

type NewMessageReceive struct {
	ChatID  int64                  `json:"Chat_Id"`
	Content *MessageContentReceive `json:"Content"`
	Token   string                 `json:"Token"`
}

type MessageContentReceive struct {
	Message   string  `json:"content"`
	Documents []int64 `json:"documents"`
	Type      string  `json:"type"`
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
	//	content, err := json.Marshal(*data.Content)
	//	if err != nil {
	//		return send, err
	//	}
	messageID, err := db.SendMessage(user.ID, data.ChatID,
		(*data.Content).Message, (*data.Content).Documents,
		db.UserMessageType, models.MessageCommandNull)
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
	var documents []models.File

	for _, v := range data.Content.Documents {
		doc, err := db.GetFile(v)
		if err != nil {
			continue
		}
		documents = append(documents, models.File{
			AuthorID: doc.AuthorID, ChatID: doc.ChatID, ID: doc.ID,
			Name: doc.Name, Path: doc.Path, RatioSize: doc.RatioSize, Size: doc.Size,
		})
	}

	var newMess models.MessageContentToUser

	newMess.Message = data.Content.Message
	newMess.Type = data.Content.Type
	newMess.Documents = &documents

	send.ID = messageID
	send.AuthorID = user.ID
	send.AuthorName = user.Name
	send.ChatID = data.ChatID
	send.Content = &newMess
	return send, nil

}

func NewMessageAnother(userQuest string) (models.NewMessageToUser, error) {
	// log.Println(*userQuest)
	var send models.NewMessageToUser
	var dataReceive struct {
		Type    string
		Content NewMessageReceive
	}

	message := []byte(userQuest)
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
	//	messageCon, err := json.Marshal(dataReceive.Content.Content.)
	//	if err != nil {
	//		return send, err
	//	}
	messageID, err := db.SendMessage(user.ID, dataReceive.Content.ChatID,
		dataReceive.Content.Content.Message, dataReceive.Content.Content.Documents,
		db.UserMessageType, models.MessageCommandNull)
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
	var documents []models.File

	for _, v := range dataReceive.Content.Content.Documents {
		doc, err := db.GetFile(v)
		if err != nil {
			continue
		}
		documents = append(documents, models.File{
			AuthorID: doc.AuthorID, ChatID: doc.ChatID,
			ID: doc.ID, Name: doc.Name, Path: doc.Path,
			RatioSize: doc.RatioSize, Size: doc.Size,
		})
	}

	var newMess models.MessageContentToUser

	newMess.Message = dataReceive.Content.Content.Message
	newMess.Type = dataReceive.Content.Content.Type
	newMess.Documents = &documents

	send.ID = messageID
	send.AuthorID = user.ID
	send.AuthorName = user.Name
	send.ChatID = dataReceive.Content.ChatID
	send.Content = &newMess
	return send, nil

}

//func NewMessagev2(msg *string)
