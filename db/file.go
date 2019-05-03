package db

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
)

const (
	CANNOT_READ_RANDOM_BYTES = "Cannot read random bytes: "
	FILE_INSERT_FAILED       = "File insertion failed: "
	CANNOT_FIND_FILE         = "Cannot find file: "
)

func CreateFile(name string, size int64, userId int64, chatId int64, ratioSize float64) (int64, string, error) {
	f := File{}
	if len(name) > 20 {
		runes := []rune(name)
		name = string(runes[len(runes)-20:])
	}
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		return -1, "", newError(CANNOT_READ_RANDOM_BYTES + err.Error())
	}
	addName := base64.URLEncoding.EncodeToString(b)

	path := strconv.FormatInt(time.Now().Unix(), 10) +
		strconv.FormatInt(size, 10) + addName + name

	f.Name = name
	f.Path = path
	f.RatioSize = ratioSize
	f.Size = size
	f.Author = &User{Id: userId}
	f.Chat = &Chat{Id: chatId}

	id, err := o.Insert(&f)
	if err != nil {
		return 0, "", newError(FILE_INSERT_FAILED + err.Error())
	}
	return id, path, nil
}

func DeleteFile(userId int64, fileId int64) (string, error) {
	f := File{Id: fileId}
	err := o.Read(&f)
	if err != nil {
		return "", newError(CANNOT_FIND_FILE + err.Error())
	}
	path := f.Path
	o.Delete(&f)
	return path, nil
}

func GetFileInformation(fileId int64) (map[string]interface{}, error) {
	final := make(map[string]interface{})
	f := File{Id: fileId}
	err := o.Read(&f)
	if err != nil {
		return final, newError(CANNOT_FIND_FILE + err.Error())
	}
	final["name"] = f.Name
	final["path"] = f.Path
	final["ratio_size"] = f.RatioSize
	final["file_id"] = f.Id
	final["size"] = f.Size
	return final, nil
}

func CheckFileRights(userId int64, fileId int64) (string, error) {
	var file File

	qb, _ := orm.NewQueryBuilder(driver)

	qb.Select("files.path").
		From("chats").
		InnerJoin("chat_users").On("chat_users.chat_id = chats.id").
		InnerJoin("users").On("users.id = chat_users.user_id").
		InnerJoin("files").On("files.chat_id = chats.id").
		Where("chat_users.chat_id = files.chat_id").
		And("users.id = chat_users.user_id").
		And("chat_users.list__invisible = 0").
		And(fmt.Sprintf("users.id = %d", userId)).
		And(fmt.Sprintf("files.id = %d", fileId))

	sql := qb.String()

	err := o.Raw(sql).QueryRow(&file)
	if err != nil {

		return "", newError(CANNOT_FIND_FILE + err.Error())
	}

	return file.Path, nil

	//f:=File{Id:fileId}
	//err:=o.Read(&f);if err!=nil{
	//	return "",err
	//}
	//userChats:= ChatUser{User:&User{Id:userId}, Chat:f.Chat}
	//err=o.Read(&userChats);if err!=nil{
	//	return "",err
	//}
	//return f.Path,nil
}
