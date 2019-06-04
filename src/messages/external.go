package messageengine

import (
	"encoding/json"
	"log"

	"github.com/alxarno/swap/src/api"

	db "github.com/alxarno/swap/db2"
	"github.com/alxarno/swap/models"
)

//ConnectActionsToDB - Bind callbacks
func ConnectActionsToDB() {
	// db.ChatCreated = chatCreated
	db.UserRequestedToChat = SendNotificationAddUserInChat
	db.SendUserMessageToSocket = SendUserMessage
	api.GetOnlineUsers = GetOnlineUsersFromSlice
}

//SendNotificationAddUserInChat - Reload only chats list on client side
func SendNotificationAddUserInChat(userID int64) {
	var message = answer{MessageType: messageTypeSystem, Result: messageSuccess, Action: messageActionUserAddedToChat}
	finish, _ := json.Marshal(message)
	for _, v := range users {
		if v.UserID == userID {
			v.SystemMessageChan <- string(finish)
		}
	}
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

//GetOnlineUsersInChat - return count online users in usersID slice
func GetOnlineUsersFromSlice(userIDs *[]int64) int64 {
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
func SendUserMessage(mID int64, chatID int64, command models.MessageCommand, authorID int64, time int64) {
	message, err := userMessageFromPure(mID, chatID, command, authorID, time)
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
// func GetKeyByToken(token string) (*rsa.PublicKey, error) {
// 	user, err := api.TestUserToken(token)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, uc := range users {
// 		if uc.UserID == user.ID {
// 			return uc.PublicKey, nil
// 		}
// 	}
// 	return nil, errors.New("User not found")
// }
