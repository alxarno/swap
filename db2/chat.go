package db2

import (
	"fmt"
	"log"
	"time"

	"github.com/alxarno/swap/models"
)

const (
	//GetUserError - cannot get user
	GetUserError = "Get user error ->"
	//InsertChatError - cannot make insert
	InsertChatError = "Insert chat error ->"
	//InsertUserInChatError - cannot insert user in chat
	InsertUserInChatError = "Insert user in chat error ->"
	//InsertUserInDialogError - cannot insert user in dialog
	InsertUserInDialogError = "Insert user in dialog error ->"
	//InsertUserInChannelError - cannot insert user in cahnnel
	InsertUserInChannelError = "Insert user in channel error ->"
	//UserAlreadyExistInChatError - cannot insert user in chat user already in
	UserAlreadyExistInChatError = "User already in chat ->"
	//AddMessageError - cannot send message
	AddMessageError = "Send message error ->"
	//GetChatError - getting chat failed
	GetChatError = "Getting chat failed ->"
	//UserIsntAuthorError - user havent rights
	UserIsntAuthorError = "User isnt author ->"
	//GetChatUserError - cannot get chat's user
	GetChatUserError = "Gettings chat user failed ->"
	//GetDeletePointsError - getting delete points failed
	GetDeletePointsError = "Failed get delete points ->"
	//SetDeletePointsError - settings delete points failed
	SetDeletePointsError = "Failed set delete points ->"
	//UpdateChatUserError - updating chat user failed
	UpdateChatUserError = "Failed update chat user ->"
	//UpdateChatError - updating chat failed
	UpdateChatError = "Failed update chat ->"
	//UserYetDidntDeleteError - user wasn't deleted
	UserYetDidntDeleteError = "User wasn't delete ->"
	//WrongChatType - got wrong chat's type
	WrongChatType = "Got wrong chat't type ->"
	//GettingUsersChatInfoFailed - getting user's chat info failed
	GettingUsersChatInfoFailed = "Getting user's chat info failed ->"
	//CheckingUserDeletedInChatFailed - checking user deleted in chat failed
	CheckingUserDeletedInChatFailed = "Checking user deleted in chat failed ->"
	//GetChatsUsersFailed - getting chat's users failed
	GetChatsUsersFailed = "Getting chat's users failed ->"
)

// ChatMode - type for decalring chat's modesChatMode
type ChatMode int

const (
	//ChatType - chat's type for Create funnction
	ChatType ChatMode = iota
	//DialogType - dialog's type for Create funnction
	DialogType
	//ChannelType - channel's type for Create funnction
	ChannelType
)

//Create - creating chat, dialog or channel and auto inserting author in it
func Create(name string, authorID int64, chattype ChatMode) (int64, error) {
	u := User{}
	if err := db.First(&u, authorID).Error; err != nil {
		return 0, DBE(GetUserError, err)
	}
	if !(chattype == ChatType || chattype == DialogType || chattype == ChannelType) {
		return 0, DBE(WrongChatType, nil)
	}
	c := Chat{Name: name, Author: u, Type: chattype}
	if err := db.Create(&c).Error; err != nil {
		return 0, DBE(InsertChatError, err)
	}
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
	chatUser.Start = time.Now().UnixNano() / 1000000
	chatUser.setDeletePoints(deletePoints)
	if err := db.Create(&chatUser).Error; err != nil {
		return DBE(InsertUserInChatError, err)
	}
	var command models.MessageCommand
	if !invited {
		switch ChatMode(chatUser.Chat.Type) {
		case ChatType:
			command = models.MessageCommandUserCreatedChat
		case DialogType:
			command = models.MessageCommandUserCreatedDialog
		case ChannelType:
			command = models.MessageCommandUserCreatedChannel
		}
	} else {
		switch ChatMode(chatUser.Chat.Type) {
		case ChatType:
			command = models.MessageCommandUserInsertedToChat
		case DialogType:
			command = models.MessageCommandUserInsertedToDialog
		case ChannelType:
			command = models.MessageCommandUserInsertedToChannel
		}
	}

	index, err := AddMessage(userID, chatID, "", []int64{}, models.SystemMessageType, command)
	if err != nil {
		return DBE(AddMessageError, err)
	}

	if SendUserMessageToSocket != nil && invited {
		SendUserMessageToSocket(
			index, chatID,
			command,
			userID,
			chatUser.Start+1)
	}
	if UserInsertedToChat != nil && invited {
		UserInsertedToChat(userID)
	}

	return nil
}

//GetChatMode - returning chat's type
func GetChatMode(chatID int64) (ChatMode, error) {
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
	if err := db.Model(&ChatUser{}).Where("chat_id = ?", chatID).Where("delete_last = 0").
		Pluck("user_id", &users).Error; err != nil {
		return &users, DBE(GetUserError, err)
	}
	return &users, nil
}

//GetChatUsersInfo - returning user's chat info by certain chat
func GetChatUsersInfo(chatID int64) (*[]models.FolkChatsInfo, error) {
	data := []models.FolkChatsInfo{}
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

//DeleteUsersInChat - delete users from certain chat or ban them(if deleteByYourself = false)
func DeleteUsersInChat(usersIDs []int64, chatID int64, deleteByYourself bool) error {
	for _, v := range usersIDs {
		chatUser := ChatUser{UserID: v, ChatID: chatID, DeleteLast: 0}
		if err := db.Where(&chatUser).Where("delete_last = ?", 0).First(&chatUser).Error; err != nil {
			log.Println(GetChatUserError, err, fmt.Sprintf("User = %d, Chat = %d", v, chatID))
			continue
		}
		deletePoints, err := chatUser.getDeletePoints()
		if err != nil {
			log.Println(GetDeletePointsError, err, fmt.Sprintf("ChatUser = %d", chatUser.ID))
			continue
		}
		dpLen := len(deletePoints)
		// If user not deleted, because delete point - [startDel, endDel], ...
		if deletePoints[dpLen-1][0] == 0 {
			messageCommand := models.MessageCommandUserWasBanned
			chatUser.Ban = true
			if deleteByYourself {
				chatUser.Ban = false
				messageCommand = models.MessageCommandUserLeaveChat
			}
			// Adding message to DB "user N was deleted/banned"
			index, err := AddMessage(v, chatID, "", []int64{}, models.SystemMessageType, messageCommand)
			if err != nil {
				return DBE(AddMessageError, err)
			}

			deletePoints[dpLen-1][0] = time.Now().UnixNano()/1000000 + 1

			chatUser.DeleteLast = deletePoints[dpLen-1][0]

			err = chatUser.setDeletePoints(deletePoints)
			if err != nil {
				log.Println(SetDeletePointsError, err,
					fmt.Sprintf("Chat User = %d, Delete Points = %#v", chatUser.ID, deletePoints))
				continue
			}
			// Send user system message that he was banned
			if UserLeaveChat != nil && !deleteByYourself {
				UserLeaveChat(v)
			}
			// Send user message to all online users that user N was deleted/banned
			if SendUserMessageToSocket != nil {
				SendUserMessageToSocket(index, chatID,
					messageCommand,
					v, chatUser.DeleteLast)
			}

			if err = db.Save(&chatUser).Error; err != nil {
				log.Println(UpdateChatUserError, err)
				continue
			}
		}
	}
	return nil
}

// CheckUserInChatDeleted - check user delete stat, if users deleted return true, else false
func CheckUserInChatDeleted(userID int64, chatID int64) (bool, error) {
	cuser := ChatUser{UserID: userID, ChatID: chatID}
	if err := db.Where(&cuser).First(&cuser).Error; err != nil {
		return false, DBE(GetChatUserError, err)
	}
	if cuser.ListInvisible || cuser.DeleteLast != 0 {
		return true, nil
	}
	return false, nil
}

// RecoveryUsersInChat - recovery users for chats if they're was deleted, but not banned
func RecoveryUsersInChat(userIDs []int64, chatID int64, recoveryByYourself bool) error {
	for _, v := range userIDs {
		c := ChatUser{UserID: v, ChatID: chatID}
		if err := db.Where(&c).Not("delete_last = ?", 0).First(&c).Error; err != nil {
			log.Println(GetChatUserError, err.Error(), fmt.Sprintf("User ID = %d, Chat ID = %d", v, chatID))
			continue
		}
		messageCommand := models.MessageCommandUserWasUnbanned
		if recoveryByYourself {
			//If user banned by another user(admin, chat's creator)
			if c.Ban {
				continue
			}
			messageCommand = models.MessageCommandUserReturnsToChat
		} else {
			c.Ban = false

		}
		deletePoints, err := c.getDeletePoints()
		if err != nil {
			log.Println(GetDeletePointsError, err, fmt.Sprintf("ChatUser = %d", c.ID))
			continue
		}
		dplen := len(deletePoints)
		// If user already deleted, because delete point - [startDel, endDel], ...
		if deletePoints[dplen-1][1] == 0 {
			deletePoints[dplen-1][1] = time.Now().UnixNano() / 1000000

			// Adding new delete point for future
			deletePoints = append(deletePoints, []int64{0, 0})
			c.DeleteLast = 0
			err := c.setDeletePoints(deletePoints)
			if err != nil {
				log.Println(SetDeletePointsError, err,
					fmt.Sprintf("Chat User = %d, Delete Points = %#v", c.ID, deletePoints))
				continue
			}

			if err = db.Save(&c).Error; err != nil {
				log.Println(UpdateChatUserError, err)
				continue
			}
			// Adding message to DB about user's return/unban
			index, err := AddMessage(v, chatID, "", []int64{}, models.SystemMessageType, messageCommand)
			if err != nil {
				return DBE(AddMessageError, err)
			}
			// Send message to online users about user's return/unban
			if SendUserMessageToSocket != nil {
				SendUserMessageToSocket(index, chatID,
					messageCommand,
					v, deletePoints[dplen-1][1])
			}
			// Send to user system message about unban
			if UserReturnToChat != nil && !recoveryByYourself {
				UserReturnToChat(v)
			}
		}
	}
	return nil
}

//GetChatSettings - return chat settings by chat's ID
func GetChatSettings(chatID int64) (models.ChatSettings, error) {
	settings := models.ChatSettings{}
	c := Chat{}
	if err := db.First(&c, chatID).Error; err != nil {
		return settings, DBE(GetChatError, err)
	}
	settings.Name = c.Name
	return settings, nil
}

//SetChatSettings - apply settigns for certain chat
func SetChatSettings(chatID int64, settings models.ChatSettings) error {
	c := Chat{}
	if err := db.First(&c, chatID).Error; err != nil {
		return DBE(GetChatError, err)
	}
	c.Name = settings.Name
	if err := db.Save(&c).Error; err != nil {
		return DBE(UpdateChatError, err)
	}
	return nil
}

//DeleteChatFromList - delete chat from certain user's menu
func DeleteChatFromList(userID int64, chatID int64) error {
	c := ChatUser{ChatID: chatID, UserID: userID}
	if err := db.Where(&c).Where("delete_last != ?", 0).First(&c).Error; err != nil {
		return DBE(GetChatUserError, err)
	}
	deleted, err := CheckUserInChatDeleted(userID, chatID)
	if err != nil {
		return DBE(CheckingUserDeletedInChatFailed, err)
	}
	if !deleted {
		return DBE(UserYetDidntDeleteError, nil)
	}
	c.ListInvisible = true
	if err := db.Save(&c).Error; err != nil {
		return DBE(UpdateChatUserError, err)
	}
	return nil
}

//GetUsersForAddByName - return users for add to certain chat if yet aren't there
func GetUsersForAddByName(chatID int64, name string) (*[]models.User, error) {
	found := []models.User{}
	existsUsersInChat, err := GetChatsUsers(chatID)
	if err != nil {
		return nil, DBE(GetChatsUsersFailed, err)
	}
	query := db.Model(&User{}).Where("id not in (?)", *existsUsersInChat).Where("(name LIKE ?", "%"+name+"%").
		Or("login LIKE ?)", "%"+name+"%")
	if err := query.Find(&found).Error; err != nil {
		return nil, DBE(GetUserError, err)
	}
	return &found, nil
}
