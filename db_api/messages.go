package db_api

import (
	"errors"
	"time"
	"github.com/Spatium-Messenger/Server/models"
	"github.com/astaxie/beego/orm"
	"fmt"
	"encoding/json"
	"github.com/AlexeyArno/Gologer"
)

var driver ="mysql"

func AddMessage(userId int64, chatId int64, content string)(int64,error){
	res,err:= CheckUserInChatDelete(userId, chatId);if err!=nil{
		return 0,err
	}
	if res{
		return 0,errors.New("user delete from chat")
	}
	m:=Message{Author:&User{Id:userId},Content:content, Chat:&Chat{Id:chatId}, Time: time.Now().Unix()}
	id,err:=o.Insert(&m);if err!=nil{
		return 0,err
	}
	return id,nil
}

func GetMessages(userId int64, chatId int64, add bool, lastIndex int64)([]*models.NewMessageToUser, error){
	type MessageTemplate struct{
		Id int64
		Content string
		AuthorId int64
		Name string
		Login string
		Time int64
	}
	templates := []MessageTemplate{}
	final := []*models.NewMessageToUser{}
	const MAX_TIME = 9999999999
	chatUser := Chat_User{User:&User{Id: userId}, Chat:&Chat{Id:chatId}}
	err:=o.Read(&chatUser);if err!=nil{
		return final,errors.New("user is not in chat")
	}

	deltimes,err := chatUser.GetDeletePoints();if err!=nil{
		return final,errors.New("cant decode delete points")
	}
	qb, _ := orm.NewQueryBuilder(driver)

	qb.Select("messages.id",
		"messages.content",
		"messages.author_id",
		"users.name",
		"users.login",
		"messages.time").
		From("messages").
		InnerJoin("users").On("messages.author_id = users.id").
		Where("messages.chat_id = ?")
	if chatUser.Chat.Type!=2{
		for i := 0; i < len(deltimes); i++ {
			if i == 0 && deltimes[0][0] == 0 {
				//messages_queries = append(messages_queries, fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", i_start, MAX_TIME))
				qb.And(fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", chatUser.Start, MAX_TIME))
			} else {
				if i == 0 {
					//messages_queries = append(messages_queries, fmt.Sprintf("((messages.time>=%d) and (messages.time<=%d)) ", i_start, r_deltimes[i][0]))
					qb.And(fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", chatUser.Start, deltimes[i][0]))
					} else if i > 0 {
					//messages_queries = append(messages_queries, fmt.Sprintf("((messages.time>=%d) and (messages.time<=%d)) ", r_deltimes[i-1][1], r_deltimes[i][0]))
					qb.And(fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", deltimes[i-1][1], deltimes[i][0]))
					if deltimes[i][0] == 0 {
						//messages_queries = append(messages_queries, fmt.Sprintf("((messages.time>=%d) and (messages.time<=%d)) ", r_deltimes[i-1][1], MAX_TIME))
						qb.And(fmt.Sprintf("((messages.time>=%d) and  (messages.time<=%d)) ", deltimes[i-1][1], MAX_TIME))
					}
				}
			}
		}
		if add {
			qb.And(fmt.Sprintf(") and ((messages.id < %d)", lastIndex))
		}
		qb.OrderBy("messages.time").Desc().Limit(80)
	}
	sql:=qb.String()
	o.Raw(sql, chatId).QueryRows(&templates)

	for _,v:= range templates{
		var Content models.MessageContentToUser
		err:= json.Unmarshal([]byte(v.Content), &Content); if err!=nil{
			Gologer.PError("Fail unmarshal : "+v.Content)
			continue
		}
		final = append(final, &models.NewMessageToUser{
			ID: v.Id,
			ChatId: chatId,
			AuthorId:v.AuthorId,
			AuthorName:v.Name,
			AuthorLogin:v.Login,
			Time:v.Time,
			Content: &Content})
	}
	return final,nil
}