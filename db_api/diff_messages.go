package db_api

import (
	"github.com/Spatium-Messenger/Server/models"
	"encoding/json"
	//"github.com/AlexeyArno/Gologer"
)

func SendMessage(UserId int64, ChatId int64, Content string, Type int)(int64,error){
	var newContent models.MessageContent
	if Type==1{
		newContent.Type = "a_msg"
	}else{
		newContent.Type = "u_msg"
	}
	newContent.Documents = []int64{}
	newContent.Message = Content

	res,err:= json.Marshal(newContent);if err!=nil{
		return -1,err
	}
	lastId,err:=addMessage(UserId,ChatId,string(res));if err!=nil{

		return -1,err
	}
	return lastId,nil
}

func SendClearMessage(UserId int64, ChatId int64, Content string)(int64,error){
	lastId,err:=addMessage(UserId,ChatId,Content);if err!=nil{
		return -1,err
	}
	return lastId,nil
}
