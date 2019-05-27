package messageengine

import (
	"encoding/json"

	db "github.com/alxarno/swap/db2"
	"github.com/alxarno/swap/models"
	"github.com/alxarno/swap/src/api"
)

type newMessageFormUser struct {
	ChatID  int64                  `json:"Chat_Id"`
	Content *models.MessageContent `json:"Content"`
	Token   string                 `json:"Token"`
}

type newMessageReceive struct {
	ChatID  int64                  `json:"chatID"`
	Content *messageContentReceive `json:"content"`
	Token   string                 `json:"token"`
}

type messageContentReceive struct {
	Message   string  `json:"content"`
	Documents []int64 `json:"documents"`
	Type      string  `json:"type"`
}

func newMessage(userQuest *string) (models.NewMessageToUser, error) {
	var send models.NewMessageToUser
	var data newMessageFormUser
	message := []byte(*userQuest)
	err := json.Unmarshal(message, &data)
	if err != nil {
		// Gologer.PError(err.Error())
		return send, err
	}

	user, err := api.TestUserToken(data.Token)
	if err != nil {
		// Gologer.PError(err.Error())
		return send, err
	}

	messageID, err := db.SendMessage(user.ID, data.ChatID,
		(*data.Content).Message, (*data.Content).Documents,
		db.UserMessageType, models.MessageCommandNull)
	if err != nil {
		return send, err
	}

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

func newMessageAnother(userQuest string) (models.NewMessageToUser, error) {
	var send models.NewMessageToUser
	var dataReceive struct {
		Type    string            `json:"mtype"`
		Content newMessageReceive `json:"content"`
	}

	message := []byte(userQuest)
	err := json.Unmarshal(message, &dataReceive)
	if err != nil {
		return send, err
	}

	user, err := api.TestUserToken(dataReceive.Content.Token)
	if err != nil {
		return send, err
	}

	messageID, err := db.SendMessage(user.ID, dataReceive.Content.ChatID,
		dataReceive.Content.Content.Message, dataReceive.Content.Content.Documents,
		db.UserMessageType, models.MessageCommandNull)
	if err != nil {
		return send, err
	}

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
	send.Type = messageTypeUser
	return send, nil

}
