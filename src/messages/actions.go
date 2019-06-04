package messageengine

import (
	"encoding/json"
	"log"

	db "github.com/alxarno/swap/db2"
	"github.com/alxarno/swap/models"
)

//ConnectActionsToDB - Bind callbacks
func ConnectActionsToDB() {
	// db.ChatCreated = chatCreated
	// db.UserRequestedToChat = requestedToChat
	db.SendUserMessageToSocket = SendUserMessage
}

//chatCreated - send notifications about chat creation
func chatCreated(AuthorID int64) {
	// chatsUsers, err := db.GetChatsUsers(chatID)
	// if err != nil {
	// 	log.Println("Error: Chat Created: GetChatUsers: ", err)
	// 	return
	// }
	// var usersOnline []int64
	// for _, b := range users {
	// 	if b.Authoriz == true {
	// 		usersOnline = append(usersOnline, b.UserId)
	// 	}
	// }

	var data = make(map[string]interface{})
	data["action"] = messageActionChatCreated
	data["type_a"] = messageTypeSystem
	data["self"] = false
	finish, _ := json.Marshal(data)
	// log.Println()
	for _, v := range users {
		// log.Println(v.UserID)
		if v.UserID == AuthorID {

			v.SystemMessageChan <- string(finish)
		}
	}

	// log.Println("Chat Created ", AuthorID, users)
}

//requestedToChat - send notifification about chat invitation
func requestedToChat(userID int64, chatID int64, command models.MessageCommand) {
	userChats, err := db.GetUsersChatsIDs(userID)
	if err != nil {
		return
	}
	var usersOnline []int64
	for _, b := range users {
		if b.Auth == true {
			usersOnline = append(usersOnline, b.UserID)
		}
	}
	notificationIds, err := db.GetOnlineUsersIDsInChat(userChats, &usersOnline)
	if err != nil {
		return
	}
	// userSettings, err := db.GetUserSettings(userID)
	// if err != nil {
	// 	return
	// }

	// userInfo,err := db.GetUser
	var data = make(map[string]interface{})
	data["action"] = messageActionUserChatInserted
	data["type_a"] = messageTypeSystem
	data["chat_id"] = chatID
	// data["command"] = command
	// data["user_name"] = userSettings["name"]
	data["self"] = false
	finish, _ := json.Marshal(data)
	// log.Println()
	for _, i := range *notificationIds {
		for _, v := range users {
			if v.UserID == i {
				if i == userID {
					data["self"] = true
					finish, _ := json.Marshal(data)
					v.SystemMessageChan <- string(finish)
					data["self"] = false
				} else {
					v.SystemMessageChan <- string(finish)
				}
			}
		}
	}
	log.Println("Request To Chat")
}

//UserMove - send notification about inc and dec online users in chats
func userMove(userID int64, moveType onlineUsersMove) {
	userChats, err := db.GetUsersChatsIDs(userID)
	if err != nil {
		return
	}
	var usersOnline []int64
	for _, b := range users {
		if b.Auth == true {
			usersOnline = append(usersOnline, b.UserID)
		}
	}
	notificationIDs, err := db.GetOnlineUsersIDsInChat(userChats, &usersOnline)
	if err != nil {
		return
	}
	var data = make(map[string]interface{})
	data["action"] = messageActionOnlineUser
	data["type"] = moveType
	data["chats"] = *userChats
	data["type_a"] = messageTypeSystem
	data["self"] = false
	finish, _ := json.Marshal(data)
	for _, i := range *notificationIDs {
		for _, v := range users {
			if v.UserID == i {
				if i == userID {
					if moveType != onlineUserDec {
						data["self"] = true
						finish, _ := json.Marshal(data)
						v.SystemMessageChan <- string(finish)
						data["self"] = false
					}
				} else {
					v.SystemMessageChan <- string(finish)
				}
			}
		}
	}
}
