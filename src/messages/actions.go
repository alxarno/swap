package messageengine

import (
	"encoding/json"

	db "github.com/alxarno/swap/db2"
)

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
	answer := struct {
		MessageType string   `json:"mtype"`
		Action      string   `json:"action"`
		Chats       *[]int64 `json:"chats"`
		Move        int      `json:"move"`
	}{
		MessageType: messageTypeSystem,
		Action:      messageActionOnlineUser,
		Chats:       userChats,
		Move:        int(moveType),
	}
	finish, _ := json.Marshal(answer)
	for _, i := range *notificationIDs {
		for _, v := range users {
			if v.UserID == i {
				v.SystemMessageChan <- string(finish)
			}
		}
	}
}
