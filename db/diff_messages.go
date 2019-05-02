package db

import (
	"encoding/json"
	"errors"

	"github.com/swap-messenger/Backend/models"
	//"github.com/AlexeyArno/Gologer"
)

func SendMessage(UserId int64, ChatId int64, Content string, Type int, command int) (int64, error) {
	var newContent models.MessageContent
	newContent.Command = command
	if Type == 1 {
		newContent.Type = "a_msg"
	} else {
		newContent.Type = "u_msg"
	}
	newContent.Documents = []int64{}
	newContent.Message = Content

	res, err := json.Marshal(newContent)
	if err != nil {
		return -1, errors.New("Marshal Error: " + err.Error())
	}
	lastID, err := addMessage(UserId, ChatId, string(res))
	if err != nil {
		return -1, errors.New("addMessage error: " + err.Error())
	}
	return lastID, nil
}

func SendClearMessage(UserId int64, ChatId int64, Content string) (int64, error) {
	lastID, err := addMessage(UserId, ChatId, Content)
	if err != nil {
		return -1, err
	}
	return lastID, nil
}
