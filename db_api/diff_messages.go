package db_api

import (
	"github.com/Spatium-Messenger/Server/models"
	"encoding/json"
)

func SendMessage(ChatId int64, UserId int64, Content string, Type int)(int64,error){
	var newContent models.MessageContentToUser
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
