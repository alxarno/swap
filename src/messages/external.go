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
	db.UserReturnToChat = sendNotificationUserReturnToChat
	db.UserLeaveChat = sendNotificationDeleteChat
	db.UserInsertedToChat = sendNotificationAddUserInChat
	db.SendUserMessageToSocket = sendUserMessage
	api.GetOnlineUsers = GetOnlineUsersFromSlice
}

// sendNotificationAddUserInChat - Reload only chats list on client side
func sendNotificationAddUserInChat(userID int64) {
	var message = answer{
		MessageType: messageTypeSystem,
		Action:      messageActionUserAddedToChat,
	}
	sendToSystem(message, userID)
}

// sendNotificationDeleteChat - Reload  chats list and now chat window close on client side
func sendNotificationDeleteChat(userID int64) {
	var message = answer{
		MessageType: messageTypeSystem,
		Action:      messageActionLeaveChat,
	}
	sendToSystem(message, userID)
}

// sendNotificationUserReturnToChat -  Reload  chats list and now chat window close on client side
func sendNotificationUserReturnToChat(userID int64) {
	var message = answer{
		MessageType: messageTypeSystem,
		Action:      messageActionReturnChat,
	}
	sendToSystem(message, userID)
}

func sendToSystem(msg interface{}, userID int64) {
	finish, _ := json.Marshal(msg)
	for _, v := range users {
		if v.UserID == userID {
			v.SystemMessageChan <- string(finish)
		}
	}
}

// GetOnlineUsersFromSlice - return count online users in usersID slice
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

// sendUserMessage - send user message to sockets
func sendUserMessage(mID int64, chatID int64, command models.MessageCommand, authorID int64, time int64) {
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
