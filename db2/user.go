package db2

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/alxarno/swap/models"
)

const (
	//UserInsertError - User insertion into DB failed
	UserInsertError = "User insert error -> "
	//UserUpdateError - User updation into DB failed
	UserUpdateError = "User updating failed -> "
	//UserWithThisLoginAlreadyExists - User with this login already exists in DB
	UserWithThisLoginAlreadyExists = "User with this login already exists -> "
	//GetChatInfoError - Getting chat's info failed
	GetChatInfoError = "Getting chat info failed -> "
	//UserNotFound - User with this data doesnt exist
	UserNotFound = "User not found -> "
	//PasswordEncodingFailed - Password encoding failed
	PasswordEncodingFailed = "Pass encoding failed -> "
	//GetMessageError - getting message failed
	GetMessageError = "Getting message failed -> "
	//MessageContentDecodeError - decoding message was failed
	MessageContentDecodeError = "Message's content decoding failed -> "
)

func encodePass(pass string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(pass))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

//GetUserByID - getting user by his ID or retrun error
func GetUserByID(ID int64) (*User, error) {
	u := User{}
	err := db.First(&u, ID).Error
	if err != nil {
		return nil, DBE(UserNotFound, err)
	}
	return &u, nil
}

//GetUserByLoginAndPass - getting user by his login and pass or return error
func GetUserByLoginAndPass(login string, pass string) (*User, error) {
	u := User{}
	encodedPass, err := encodePass(pass)
	if err != nil {
		return nil, DBE(PasswordEncodingFailed, err)
	}
	err = db.Where("login = ?", login).
		Where("pass = ?", encodedPass).
		First(&u).Error
	if err != nil {
		return nil, DBE(UserNotFound, err)
	}
	return &u, nil
}

//CreateUser - creating creating user record in DB by his login, Pass and name
//pass will be encoding to sha256
func CreateUser(login string, pass string, name string) (int64, error) {
	u := User{}

	if !db.Where("login = ?", login).First(&u).RecordNotFound() {
		return 0, DBE(UserWithThisLoginAlreadyExists, nil)
	}
	encodedPass, err := encodePass(pass)
	if err != nil {
		return 0, DBE(PasswordEncodingFailed, err)
	}
	u.Pass = encodedPass
	u.Name = name
	u.Login = login
	if err := db.Create(&u).Error; err != nil {
		return 0, DBE(UserInsertError, err)
	}
	return u.ID, nil
}

//GetUserChats - return users chats
func GetUserChats(userID int64) (*[]models.UserChatInfo, error) {
	info := []chatInfo{}
	mes := Message{}
	res := []models.UserChatInfo{}
	//Query for chat's info
	query := db.Table("chat_users").
		Select(`chats.id, chats.name, chats.author_id,
		chats.type, chat_users.delete_last,chat_users.ban`).
		Joins("inner join chats on chat_users.chat_id = chats.id").
		Where("list_invisible = ?", 0).
		Where("user_id = ?", userID)
	if err := query.Scan(&info).Error; err != nil {
		return nil, DBE(GetChatInfoError, err)
	}
	//Query for last messages
	query = db.Table("messages").
		Joins("inner join users on messages.author_id = users.id")
	msgContent := models.MessageContent{}
	for _, v := range info {
		if err := query.Where("messages.chat_id=?", v.ID).Last(&mes).Error; err != nil {
			return nil, DBE(GetMessageError, err)
		}
		// log.Println(mes, v.ID)
		err := json.Unmarshal([]byte(mes.Content), &msgContent)
		if err != nil {
			return nil, DBE(MessageContentDecodeError, err)
		}
		deleted := true
		if !v.Ban && v.DeleteLast == 0 {
			deleted = false
		}
		res = append(res, models.UserChatInfo{
			ID: v.ID, Name: v.Name, Type: v.Type,
			LastSender: mes.Author.Name, AdminID: v.AuthorID,
			LastMessage: &msgContent, LastMessageTime: mes.Time,
			View: 0, Delete: deleted, Online: 0})
	}
	return &res, nil
}

//GetUsersChatsIDs - return user's chats ID
func GetUsersChatsIDs(userID int64) (*[]int64, error) {
	IDs := []int64{}
	if err := db.Model(&ChatUser{}).
		Where("user_id = ?", userID).
		Pluck("chat_id", &IDs).Error; err != nil {
		return nil, DBE(GetChatUserError, err)
	}
	return &IDs, nil
}

//GetOnlineUsersIDsInChat - return online users IDs in chats by online usesrs slice
func GetOnlineUsersIDsInChat(chatsID *[]int64, usersOnlineID *[]int64) (*[]int64, error) {
	res := []int64{}
	query := db.Model(&ChatUser{}).Where("user_id", usersOnlineID).
		Where("chat_id", chatsID).Where("ban = ?", 0).Where("list_invisible = 0").
		Where("delete_last = ?", 0)
	if err := query.Pluck("user_id", &res).Error; err != nil {
		return nil, DBE(GetChatUserError, err)
	}
	return &res, nil
}

//GetUserSettings - return user's settings
func GetUserSettings(userID int64) (models.UserSettings, error) {
	user := User{}
	settings := models.UserSettings{}
	if err := db.First(&user, userID).Error; err != nil {
		return settings, DBE(UserNotFound, err)
	}
	settings.Name = user.Name
	return settings, nil
}

//SetUserSettigns - save settings to DB
func SetUserSettigns(userID int64, settings models.UserSettings) error {
	user := User{}
	if err := db.First(&user, userID).Error; err != nil {
		return DBE(UserNotFound, err)
	}
	user.Name = settings.Name
	if err := db.Save(&user).Error; err != nil {
		return DBE(UserUpdateError, err)
	}
	return nil
}
