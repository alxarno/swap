package message_engine

import (
	"encoding/json"
	"log"

	db "github.com/swap-messenger/swap/db2"
	"github.com/swap-messenger/swap/models"
)

func ConnectActionsToDB() {
	db.ChatCreated = ChatCreated
	db.UserRequestedToChat = RequestedToChat
}

func ChatCreated(AuthorId int64) {
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
	data["action"] = models.MessageActionChatCreated
	data["type_a"] = models.MessageActionTypeSystem
	data["self"] = false
	finish, _ := json.Marshal(data)
	// log.Println()
	for _, v := range users {
		log.Println(v.UserId)
		if v.UserId == AuthorId {

			v.SystemMessChan <- string(finish)
		}
	}

	log.Println("Chat Created ", AuthorId, users)
}

func RequestedToChat(userID int64, chatID int64, command models.MessageCommand) {
	userChats, err := db.GetUsersChatsIDs(userID)
	if err != nil {
		return
	}
	var usersOnline []int64
	for _, b := range users {
		if b.Authoriz == true {
			usersOnline = append(usersOnline, b.UserId)
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
	data["action"] = models.MessageActionUserChatInserted
	data["type_a"] = models.MessageActionTypeSystem
	data["chat_id"] = chatID
	// data["command"] = command
	// data["user_name"] = userSettings["name"]
	data["self"] = false
	finish, _ := json.Marshal(data)
	// log.Println()
	for _, i := range *notificationIds {
		for _, v := range users {
			if v.UserId == i {
				if i == userID {
					data["self"] = true
					finish, _ := json.Marshal(data)
					v.SystemMessChan <- string(finish)
					data["self"] = false
				} else {
					v.SystemMessChan <- string(finish)
				}
			}
		}
	}
	log.Println("Request To Chat")
}

func UserMove(userId int64, mType string) {
	userChats, err := db.GetUsersChatsIDs(userId)
	if err != nil {
		return
	}
	var usersOnline []int64
	for _, b := range users {
		if b.Authoriz == true {
			usersOnline = append(usersOnline, b.UserId)
		}
	}
	notificationIDs, err := db.GetOnlineUsersIDsInChat(userChats, &usersOnline)
	if err != nil {
		return
	}
	var data = make(map[string]interface{})
	data["action"] = models.MessageActionOnlineUser
	data["type"] = mType
	data["chats"] = *userChats
	data["type_a"] = models.MessageActionTypeSystem
	data["self"] = false
	finish, _ := json.Marshal(data)
	for _, i := range *notificationIDs {
		for _, v := range users {
			if v.UserId == i {
				if i == userId {
					if mType != "-" {
						data["self"] = true
						finish, _ := json.Marshal(data)
						v.SystemMessChan <- string(finish)
						data["self"] = false
					}
				} else {
					v.SystemMessChan <- string(finish)
				}
			}
		}
	}
}
