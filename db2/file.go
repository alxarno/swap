package db2

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"
)

const (
	//CannotReadRandomBytes - reading reandom bytes failed
	CannotReadRandomBytes = "Reading random bytes failed: "
	//FileInsertingFailed = inserting file into db failed
	FileInsertingFailed = "Inserting file into db failed: "
	//CannotFindFile = file not found
	CannotFindFile = "File not found: "
	//CannotDeleteFile = file deleting failed
	CannotDeleteFile = "File deleting failed: "
)

func getRandomString(len int) (string, error) {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		return "", DBE(CannotReadRandomBytes, err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

//CreateFile - insert file data into db
func CreateFile(name string, size int64, userID int64,
	chatID int64, ratio float64) (int64, string, error) {

	deleted, err := CheckUserInChatDeleted(userID, chatID)
	if err != nil {
		return 0, "", DBE(CheckingUserInChatFailed, err)
	}
	if deleted {
		return 0, "", DBE(UserDeletedFromChat, nil)
	}

	if len(name) > 20 {
		runes := []rune(name)
		name = string(runes[len(runes)-20:])
	}
	additional, err := getRandomString(20)
	if err != nil {
		return 0, "", err
	}
	path := strconv.FormatInt(time.Now().Unix(), 10) +
		strconv.FormatInt(size, 10) + additional + name

	f := File{
		Name: name, Path: path, RatioSize: ratio, Size: size,
		AuthorID: userID, ChatID: chatID,
	}
	if err := db.Create(&f).Error; err != nil {
		return f.ID, f.Path, DBE(FileInsertingFailed, err)
	}
	return f.ID, f.Path, nil
}

//DeleteFile - delete file info into db and return path to file
func DeleteFile(fileID int64) (string, error) {
	f := File{ID: fileID}
	if err := db.Where(&f).First(&f).Error; err != nil {
		return "", DBE(CannotFindFile, err)
	}
	path := f.Path
	if err := db.Delete(&f).Error; err != nil {
		return "", DBE(CannotDeleteFile, err)
	}
	return path, nil
}

//GetFile - return pointer to file
func GetFile(fileID int64) (*File, error) {
	f := File{ID: fileID}
	if err := db.Where(&f).First(&f).Error; err != nil {
		return nil, DBE(CannotFindFile, err)
	}
	return &f, nil
}

//CheckFileRights - checking if user have rights for getting file
//(e.g. user is in a chat where file is), return path to file
func CheckFileRights(userID int64, fileID int64) (string, error) {
	path := ""
	query := db.Model(&Chat{}).
		Joins("inner join chat_users on chat_users.chat_id = chats.id").
		Joins("inner join users on users.id = chats_users.user_id").
		Joins("inner join files on files.chat_id = chats.id").
		Where("chat_users.chat_id = files.chat_id").
		Where("users.id = chat_users.user_id").
		Where("chat_users.list__invisible = ?", 0).
		Where("chat_users.last_deleted = ?", 0).
		Where("users.id = ?", userID).
		Where("files.id = ?", fileID)
	if err := query.Pluck("files.path", &path).Error; err != nil {
		return path, DBE(CannotFindFile, err)
	}
	return path, nil
}
