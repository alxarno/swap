package db_api

import (
	"encoding/base64"
	"time"
	"strconv"
	"crypto/rand"
)

func CreateFile(name string,size int64,userId int64, chatId int64, ratioSize float64)(int64, string,error){
	f:= File{}
	if len(name)>20{
		runes := []rune(name)
		name = string(runes[len(runes)-20:])
	}
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		return -1,"",err
	}
	addName := base64.URLEncoding.EncodeToString(b)

	path :=  strconv.FormatInt(time.Now().Unix(),10)+
		strconv.FormatInt(size,10)+addName+name

	f.Name = name
	f.Path = path
	f.RatioSize = ratioSize
	f.Size = size
	f.Author=&User{Id:userId}
	f.Chat=&Chat{Id:chatId}

	id,err:= o.Insert(&f);if err!=nil{
		return 0,"",err
	}
	return id,path,nil
}

func DeleteFile(userId int64, fileId int64)(string,error){
	f:= File{Id: fileId}
	err:=o.Read(&f);if err!=nil{
		return "",err
	}
	path := f.Path
	o.Delete(&f)
	return path,nil
}

func GetFileInformation(fileId int64)(map[string]interface{},error){
	final := make(map[string]interface{})
	f:=File{Id: fileId}
	err:=o.Read(&f);if err!=nil{
		return final,err
	}
	final["name"] = f.Name
	final["path"] = f.Path
	final["ratio_size"] = f.RatioSize
	final["file_id"] = f.Id
	final["size"] = f.Size
	return final, nil
}

func CheckFileRights(userId int64, fileId int64)(string,error){
	f:=File{Id:fileId}
	err:=o.Read(&f);if err!=nil{
		return "",err
	}
	userChats:= ChatUser{User:&User{Id:userId}, Chat:f.Chat}
	err=o.Read(&userChats);if err!=nil{
		return "",err
	}
	return f.Path,nil
}