package messageengine

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"log"

	"github.com/alxarno/swap/models"
	"github.com/alxarno/swap/src/api"
)

//SendNotificationAddUserInChat - Reload only chats list on client side
func SendNotificationAddUserInChat(userID int64) error {
	var message = answer{
		MessageType: messageTypeSystem,
		Action:      messageActionUserAddedToChat,
	}
	finish, _ := json.Marshal(message)
	for _, v := range users {
		if v.UserID == userID {
			v.SystemMessageChan <- string(finish)
		}
	}
	return nil
}

//SendNotificationDeleteChat - Reload  chats list and now chat window close on client side
func SendNotificationDeleteChat(userID int64) error {
	var message = answer{
		MessageType: messageTypeSystem,
		Action:      messageActionDeleteChat,
	}
	finish, _ := json.Marshal(message)
	for _, v := range users {
		if v.UserID == userID {
			v.SystemMessageChan <- string(finish)
		}
	}
	return nil
}

//GetOnlineUsersInChat - return count online users in certain chat
func GetOnlineUsersInChat(userIDs *[]int64) int64 {
	var count int64
	count = 0
	for _, v := range users {
		for _, b := range *userIDs {
			if v.UserID == b {
				count++
			}
		}
	}
	return count
}

// SendUserMessage - send user message to sockets
func SendUserMessage(mID int64, chatID int64, content *models.MessageContentToUser, authorID int64, time int64) {
	message, err := userMessageFromPure(mID, chatID, content, authorID, time)
	if err != nil {
		log.Println("SendUSerMessage extrenal.go -> ", err.Error())
		return
	}

	sendMessages <- message
}

//SendMessage - send message
func SendMessage(msg models.NewMessageToUser) {
	sendMessages <- msg
}

//SendForceMessage - send force message
func SendForceMessage(msg models.ForceMsgToUser) {
	forceSendMessages <- msg
}

//GetKeyByToken - return public key for user
func GetKeyByToken(token string) (*rsa.PublicKey, error) {
	user, err := api.TestUserToken(token)
	if err != nil {
		return nil, err
	}
	for _, uc := range users {
		if uc.UserID == user.ID {
			return uc.PublicKey, nil
		}
	}
	return nil, errors.New("User not found")
}
