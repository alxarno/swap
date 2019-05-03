package db

import (
	"encoding/json"

	"github.com/swap-messenger/swap/models"
	//"github.com/AlexeyArno/Gologer"
)

const (
	SYSTEM_MESSAGE_TYPE = "a_msg"
	USER_MESSAGE_TYPE   = "u_msg"
)

const (
	MARSHAL_ERRROR        = "Marshaling failed: "
	ADDING_MESSAGE_FAILED = "Adding message failed: "
)

func SendMessage(UserId int64, ChatId int64, Content string, Type int, command int) (int64, error) {
	var newContent models.MessageContent
	newContent.Command = command
	if Type == 1 {
		newContent.Type = SYSTEM_MESSAGE_TYPE
	} else {
		newContent.Type = USER_MESSAGE_TYPE
	}
	newContent.Documents = []int64{}
	newContent.Message = Content

	res, err := json.Marshal(newContent)
	if err != nil {
		return -1, newError(MARSHAL_ERRROR + err.Error())
	}
	lastID, err := addMessage(UserId, ChatId, string(res))
	if err != nil {
		return -1, newError(ADDING_MESSAGE_FAILED + err.Error())
	}
	return lastID, nil
}

func SendClearMessage(UserId int64, ChatId int64, Content string) (int64, error) {
	lastID, err := addMessage(UserId, ChatId, Content)
	if err != nil {
		return -1, newError(ADDING_MESSAGE_FAILED + err.Error())
	}
	return lastID, nil
}
