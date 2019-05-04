package db2

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/swap-messenger/swap/models"
)

const (
	//UserInsertError - User insertion into DB failed
	UserInsertError = "User insert error: "
	//UserUpdateError - User updation into DB failed
	UserUpdateError = "User updating failed: "
	//UserWithThisLoginAlreadyExists - User with this login already exists in DB
	UserWithThisLoginAlreadyExists = "User with this login already exists: "
	//GetChatInfoError - Getting chat's info failed
	GetChatInfoError = "Getting chat info failed: "
	//UserNotFound - User with this data doesnt exist
	UserNotFound = "User not found: "
	//PasswordEncodingFailed - Password encoding failed
	PasswordEncodingFailed = "Pass encoding failed"
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
	tx := db.Begin()
	if err := tx.Create(&u).Error; err != nil {
		tx.Rollback()
		return 0, DBE(UserInsertError, err)
	}
	tx.Commit()
	return u.ID, nil
}

//GetUserChats - return users chats
func GetUserChats(userID int64) ([]*models.UserChatInfo, error) {
	//
	return []*models.UserChatInfo{}, nil
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
