package db2

import (
	"encoding/json"
	"time"

	"github.com/alxarno/swap/models"
)

//MessageType - type for message type's aliases
type MessageType string

const (
	//SystemMessageType - system message alias - "a_msg"
	SystemMessageType MessageType = "a_msg"
	//UserMessageType - user message alias - "u_msg"
	UserMessageType      MessageType = "u_msg"
	messsageTrancheLimit             = 80
)

const (
	//MarshalingFailed - marshaling failed
	MarshalingFailed = "Marshaling failed: "
	//UnmarshalingFailed - unmarshaling failed
	UnmarshalingFailed = "Unmarshaling failed: "
	//AddingMessageFailed - inserting message failed
	AddingMessageFailed = "Adding message failed: "
	//CheckingUserInChatFailed - checking user in chat failed
	CheckingUserInChatFailed = "Checking user in chat faile: "
	//UserDeletedFromChat - user was deleted from chat
	UserDeletedFromChat = "User deleted from chat: "
	//MessageInsertingFailed - message inserting was failed
	MessageInsertingFailed = "Message insert error: "
	//GettingFileInfoFailed - cannot get file's info
	GettingFileInfoFailed = "Getting file's information failed: "
)

//addMessage - inserting message into table
func addMessage(userID int64, chatID int64, content string) (int64, error) {
	deleted, err := CheckUserInChatDeleted(userID, chatID)
	if err != nil {
		return 0, DBE(CheckingUserInChatFailed, err)
	}
	if deleted {
		return 0, DBE(UserDeletedFromChat, nil)
	}
	m := Message{
		AuthorID: userID,
		Content:  content,
		ChatID:   chatID,
		Time:     time.Now().Unix(),
	}
	if err := db.Create(&m).Error; err != nil {
		return 0, DBE(MessageInsertingFailed, err)
	}
	return m.ID, nil
}

//GetMessages - retrun user's messages in certain chat, supporting tranches(pages)
func GetMessages(userID int64, chatID int64, tranches bool, lastID int64) (*[]models.NewMessageToUser, error) {
	templates := []messageTemplate{}
	response := []models.NewMessageToUser{}
	chatUser := ChatUser{UserID: userID, ChatID: chatID}
	if err := db.Where(&chatUser).First(&chatUser).Error; err != nil {
		return nil, DBE(GetChatUserError, err)
	}
	deletePoints, err := chatUser.getDeletePoints()
	if err != nil {
		return nil, DBE(GetDeletePointsError, err)
	}
	query := db.Model(&Message{}).
		Select("messages.id, messages.content, users.name, users.login, messages.time").
		Joins("INNER JOIN users ON messages.author_id = users.id").
		Where("messages.chat_id = ?", chatID)
	if ChatMode(chatUser.Chat.Type) != ChannelType {
		for i := 0; i < len(deletePoints); i++ {
			// User never as deleted
			if i == 0 && deletePoints[0][0] == 0 {
				query = query.Where("messages.time >= ?", chatUser.Start)
			} else {
				if i == 0 {
					//From chat joined to first delete date
					query = query.Where("messages.time >= ?", chatUser.Start).
						Where("messages.time<?", deletePoints[i][0])
				} else {
					query = query.Where("messages.time >= ?", deletePoints[i-1][1]).
						Where("messages.time <= ?", deletePoints[i][0])
					if deletePoints[i][0] == 0 {
						query = query.Where("messages.time >= ?", deletePoints[i-1][1])
					}
				}
			}
		}
		if tranches {
			query = query.Where("messages.id > ?", lastID)
		}
		query = query.Order("messages.time asc").Limit(messsageTrancheLimit)
	}

	if err := query.Scan(&templates).Error; err != nil {
		return nil, DBE(GetMessageError, err)
	}

	for _, v := range templates {
		content := models.MessageContent{}
		if err := json.Unmarshal([]byte(v.Content), &content); err != nil {
			return nil, DBE(UnmarshalingFailed, err)
		}
		docs := []models.File{}
		for _, d := range content.Documents {
			doc, err := GetFile(d)
			if err != nil {
				return nil, DBE(GettingFileInfoFailed, err)
			}
			docs = append(docs, models.File{
				ID: doc.ID, AuthorID: doc.AuthorID,
				ChatID: doc.ChatID, Size: doc.Size,
				Duration: doc.Duration,
				Name:     doc.Name, Path: doc.Path,
				RatioSize: doc.RatioSize,
			})
		}
		mes := models.MessageContentToUser{
			Documents: &docs, Message: content.Message,
			Type: content.Type, Command: content.Command,
		}

		response = append(response, models.NewMessageToUser{
			ID: v.ID, ChatID: chatID, AuthorID: v.AuthorID,
			AuthorLogin: v.Login, AuthorName: v.Name,
			Time: v.Time, Content: &mes})
	}
	return &response, nil
}

//SendMessage - handle inserting message into db
func SendMessage(userID int64, chatID int64, content string, docs []int64,
	mtype MessageType, command models.MessageCommand) (int64, error) {
	mcontent := models.MessageContent{
		Command: int(command), Type: string(mtype),
		Documents: docs, Message: content,
	}

	jcontent, err := json.Marshal(mcontent)
	if err != nil {
		return 0, DBE(MarshalingFailed, err)
	}
	lastID, err := addMessage(userID, chatID, string(jcontent))
	if err != nil {
		return 0, DBE(AddingMessageFailed, err)
	}
	return lastID, nil
}
