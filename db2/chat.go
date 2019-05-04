package db2

import (
	"fmt"
	"log"
	"time"

	"github.com/swap-messenger/swap/models"
)

const (
	//GetUserError - cannot get user
	GetUserError = "Get user error: "
	//InsertChatError - cannot make insert
	InsertChatError = "Insert chat error: "
	//InsertUserInChatError - cannot insert user in chat
	InsertUserInChatError = "Insert user in chat error: "
	//InsertUserInDialogError - cannot insert user in dialog
	InsertUserInDialogError = "Insert user in dialog error: "
	//InsertUserInChannelError - cannot insert user in cahnnel
	InsertUserInChannelError = "Insert user in channel error: "
	//UserAlreadyExistInChatError - cannot insert user in chat user already in
	UserAlreadyExistInChatError = "User already in chat: "
	//SendMessageError - cannot send message
	SendMessageError = "Send message error: "
	//GetChatError - getting chat failed
	GetChatError = "Getting chat failed: "
	//UserIsntAuthorError - user havent rights
	UserIsntAuthorError = "User isnt author: "
	//GetChatUserError - cannot get chat's user
	GetChatUserError = "Gettings chat user failed: "
	//GetDeletePointsError - getting delete points failed
	GetDeletePointsError = "Failed get delete points: "
	//SetDeletePointsError - settings delete points failed
	SetDeletePointsError = "Failed set delete points: "
	//UpdateChatUserError - updating chat user failed
	UpdateChatUserError = "Failed update chat user: "
	//UpdateChatError - updating chat failed
	UpdateChatError = "Failed update chat: "
	//UserYetDidntDeleteError - user wasn't deleted
	UserYetDidntDeleteError = "User wasn't delete: "
	//WrongChatType - got wrong chat's type
	WrongChatType = "Got wrong chat't type: "
	//GettingUsersChatInfoFailed - getting user's chat info failed
	GettingUsersChatInfoFailed = "Getting user's chat info failed: "
)

const (
	//ChatType - chat's type for Create funnction
	ChatType = iota
	//DialogType - dialog's type for Create funnction
	DialogType
	//ChannelType - channel's type for Create funnction
	ChannelType
)

//Create - creating chat, dialog or channel
func Create(name string, authorID int64, chattype int) (int64, error) {
	u := User{}
	if err := db.First(&u, authorID).Error; err != nil {
		return 0, DBE(GetUserError, err)
	}
	if !(chattype == ChatType || chattype == DialogType || chattype == ChannelType) {
		return 0, DBE(WrongChatType, nil)
	}
	tx := db.Begin()
	c := Chat{Name: name, Author: u, Type: chattype}
	if err := tx.Create(&c).Error; err != nil {
		tx.Rollback()
		return 0, DBE(InsertChatError, err)
	}
	tx.Commit()
	err := InsertUserInChat(u.ID, c.ID, false)
	if err != nil {
		switch chattype {
		case ChatType:
			return c.ID, DBE(InsertUserInChatError, err)
		case DialogType:
			return c.ID, DBE(InsertUserInDialogError, err)
		case ChannelType:
			return c.ID, DBE(InsertUserInChannelError, err)
		}
	}
	if ChatCreated != nil {
		ChatCreated(authorID)
	}
	return c.ID, nil
}

//InsertUserInChat - adding user to chat
func InsertUserInChat(userID int64, chatID int64, invited bool) error {
	chatUser := ChatUser{ChatID: chatID, UserID: userID}
	if !db.Where(&chatUser).First(&chatUser).RecordNotFound() {
		return DBE(UserAlreadyExistInChatError, nil)
	}
	//Creating new delete points
	var deletePoints [][]int64
	deletePoints = append(deletePoints, []int64{0, 0})
	chatUser.Start = time.Now().Unix()
	chatUser.SetDeletePoints(deletePoints)
	if err := db.Create(&chatUser).Error; err != nil {
		return DBE(InsertUserInChatError, err)
	}
	var command int
	if !invited {
		switch chatUser.Chat.Type {
		case ChatType:
			command = models.MessageCommandUserCreatedChat
		case DialogType:
			command = models.MessageCommandUserCreatedDialog
		case ChannelType:
			command = models.MessageCommandUserCreatedChannel
		}
	} else {
		switch chatUser.Chat.Type {
		case ChatType:
			command = models.MessageCommandUserInsertedToChat
		case DialogType:
			command = models.MessageCommandUserInsertedToDialog
		case ChannelType:
			command = models.MessageCommandUserInsertedToChannel
		}
		if UserRequestedToChat != nil {
			UserRequestedToChat(userID, chatID, command)
		}
	}
	// _, err = SendMessage(UserId, ChatId, content, 1, command)
	// if err != nil {
	// 	return DBE(SendMessageError, err)
	// }

	return nil
}

//GetChatType - returning chat's type
func GetChatType(chatID int64) (int, error) {
	chat := Chat{}
	if err := db.First(&chat, chatID).Error; err != nil {
		return 0, DBE(GetChatError, err)
	}
	return chat.Type, nil
}

//CheckUserRights - return is user author of the chat (error - no, nil - yes)
func CheckUserRights(userID int64, chatID int64) error {
	chat := Chat{}
	if err := db.First(&chat, chatID).Error; err != nil {
		return DBE(GetChatError, err)
	}
	if chat.AuthorID != userID {
		return DBE(UserIsntAuthorError, nil)
	}
	return nil
}

//GetChatsUsers - returning user's ids in the certain chat
func GetChatsUsers(chatID int64) (*[]int64, error) {
	users := []int64{}
	if err := db.Find(&User{}).Pluck("ID", &users).Error; err != nil {
		return &users, DBE(GetUserError, err)
	}
	return &users, nil
}

//GetChatUsersInfo - returning user's chat info by certain chat
func GetChatUsersInfo(chatID int64) (*[]models.UserChatsInfo, error) {
	data := []models.UserChatsInfo{}
	err := db.Table("chat_users").
		Select("users.id, users.login,users.name,chat_users.delete_last,chat_users.ban").
		Joins("inner join users on  chat_users.user_id = users.id").
		Where("chat_users.chat_id = ?", chatID).
		Where("chat_users.list_invisible = 0").Scan(&data).Error
	if err != nil {
		return nil, DBE(GettingUsersChatInfoFailed, err)
	}
	for i := range data {
		if data[i].DeleteLast != 0 {
			data[i].DeleteLast = 1
		}
	}
	return &data, nil
}

//DeleteUsersInChat - delete users from certain chat or ban them
func DeleteUsersInChat(usersIDs []int64, chatID int64, deleteByYourself bool) error {
	for _, v := range usersIDs {
		chatUser := ChatUser{UserID: v, ChatID: chatID, DeleteLast: 0}
		if err := db.Where(&chatUser).Where("delete_last = ?", 0).First(&chatUser).Error; err != nil {
			log.Println(GetChatUserError, err, fmt.Sprintf("User = %d, Chat = %d", v, chatID))
			continue
		}
		deletePoints, err := chatUser.GetDeletePoints()
		if err != nil {
			log.Println(GetDeletePointsError, err, fmt.Sprintf("ChatUser = %d", chatUser.ID))
			continue
		}
		dpLen := len(deletePoints)
		// If user not deleted, because delete point - [startDel, endDel], ...
		if deletePoints[dpLen-1][0] == 0 {
			deletePoints[dpLen-1][0] = time.Now().Unix()
			chatUser.DeleteLast = deletePoints[dpLen-1][0]
			chatUser.Ban = true
			if deleteByYourself {
				chatUser.Ban = false
			}
			err := chatUser.SetDeletePoints(deletePoints)
			if err != nil {
				log.Println(SetDeletePointsError, err,
					fmt.Sprintf("Chat User = %d, Delete Points = %#v", chatUser.ID, deletePoints))
				continue
			}
			if err = db.Save(&chatUser).Error; err != nil {
				log.Println(UpdateChatUserError, err)
				continue
			}
		}
	}
	return nil
}

//RecoveryUsersInChat - recovery users for chats if they're was deleted, but not banned
func RecoveryUsersInChat(userIDs []int64, chatID int64, recoveryByYourself bool) error {
	for _, v := range userIDs {
		c := ChatUser{UserID: v, ChatID: chatID}
		if err := db.Where(&c).Not("delete_last = ?", 0).First(&c).Error; err != nil {
			log.Println(GetChatUserError, err.Error(), fmt.Sprintf("User ID = %d, Chat ID = %d", v, chatID))
			continue
		}
		if recoveryByYourself {
			//If user banned by another user(admin, chat's creator)
			if c.Ban {
				continue
			}
		} else {
			c.Ban = false
		}
		deletePoints, err := c.GetDeletePoints()
		if err != nil {
			log.Println(GetDeletePointsError, err, fmt.Sprintf("ChatUser = %d", c.ID))
			continue
		}
		dplen := len(deletePoints)
		// If user already deleted, because delete point - [startDel, endDel], ...
		if deletePoints[dplen-1][1] == 0 {
			deletePoints[dplen-1][1] = time.Now().Unix()
			//Adding new delete point for future
			deletePoints = append(deletePoints, []int64{0, 0})
			c.DeleteLast = 0
			err := c.SetDeletePoints(deletePoints)
			if err != nil {
				log.Println(SetDeletePointsError, err,
					fmt.Sprintf("Chat User = %d, Delete Points = %#v", c.ID, deletePoints))
				continue
			}
			if err = db.Save(&c).Error; err != nil {
				log.Println(UpdateChatUserError, err)
				continue
			}
		}
	}
	return nil
}

// func GetChat
