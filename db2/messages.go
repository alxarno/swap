package db2

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alxarno/swap/models"
)

const (
	messsageTrancheLimit = 80
)

const (
	//MarshalingFailed - marshaling failed
	MarshalingFailed = "Marshaling failed ->"
	//UnmarshalingFailed - unmarshaling failed
	UnmarshalingFailed = "Unmarshaling failed ->"
	//AddingMessageFailed - inserting message failed
	AddingMessageFailed = "Adding message failed ->"
	//CheckingUserInChatFailed - checking user in chat failed
	CheckingUserInChatFailed = "Checking user in chat faile ->"
	//UserDeletedFromChat - user was deleted from chat
	UserDeletedFromChat = "User deleted from chat ->"
	//MessageInsertingFailed - message inserting was failed
	MessageInsertingFailed = "Message insert error ->"
	//GettingFileInfoFailed - cannot get file's info
	GettingFileInfoFailed = "Getting file's information failed ->"
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
		Time:     time.Now().UnixNano() / 1000000,
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
		subquery := ""
		for i := 0; i < len(deletePoints); i++ {
			// User never as deleted
			if i == 0 && deletePoints[0][0] == 0 {
				// query = query.Where("messages.time >= ?", chatUser.Start)
				subquery += fmt.Sprintf(" messages.time >= %d", chatUser.Start)
			} else {
				if i == 0 {
					//From chat joined to first delete date
					// query = query.Where("( messages.time >= ?", chatUser.Start).
					// Where("messages.time<? )", deletePoints[i][0])
					subquery += fmt.Sprintf(" ((messages.time >= %d) AND (messages.time < %d))", chatUser.Start, deletePoints[i][0])
				} else {
					// query = query.Or("( messages.time >= ?", deletePoints[i-1][1]).
					// Where("messages.time <= ? )", deletePoints[i][0])
					subquery += fmt.Sprintf(" OR ((messages.time >= %d) AND (messages.time <= %d))", deletePoints[i-1][1], deletePoints[i][0])
					if deletePoints[i][0] == 0 {
						// query = query.Where("messages.time >= ?", deletePoints[i-1][1])
						subquery += fmt.Sprintf(" OR (messages.time >= %d)", deletePoints[i-1][1])
					}
				}
			}
		}
		query = query.Where(subquery)
		// fmt.Println(subquery)
		if tranches {
			query = query.Where("messages.id > ?", lastID)
		}
		query = query.Order("messages.time asc").Limit(messsageTrancheLimit)
	}

	if err := query.Scan(&templates).Error; err != nil {
		return nil, DBE(GetMessageError, err)
	}

	for _, v := range templates {
		// log.Println(v.Content)
		content := models.MessageContent{}
		if err := json.Unmarshal([]byte(v.Content), &content); err != nil {
			return nil, DBE(UnmarshalingFailed, err)
		}
		docs := []models.File{}
		for _, d := range content.Documents {
			doc, err := GetFile(d)
			// log.Print(doc, d)
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

//AddMessage - handle inserting message into db
func AddMessage(userID int64, chatID int64, content string, docs []int64,
	mtype models.MessageType, command models.MessageCommand) (int64, error) {
	mcontent := models.MessageContent{
		Command: int(command), Type: int(mtype),
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
