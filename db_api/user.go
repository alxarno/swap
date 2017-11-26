package db_api

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/astaxie/beego/orm"
	"github.com/AlexeyArno/Gologer"
	"errors"
	"github.com/Spatium-Messenger/Server/models"
	"encoding/json"
	"strconv"
	"strings"
)
var Driver = "mysql"

//import "fmt
//import "github.com/Spatium-Messenger/Server/src/api/methods"

func GetUser(s_type string, data map[string]interface{})(*User, error){
	if s_type =="login"{
		h := sha256.New()
		h.Write([]byte(data["pass"].(string)))
		u := User{
			Login: data["login"].(string),
			Pass: base64.StdEncoding.EncodeToString(h.Sum(nil))}
		err := o.Read(&u)
		if err!=nil{
			return nil, err
		}
		return &u,nil
	}else{
		u := User{Id: data["id"].(int64)}
		err := o.Read(&u)
		if err!=nil{
			return nil, err
		}
		return &u,nil
	}
}

func CreateUser(login string, pass string, u_name string)(int64, error){
	u:= User{Login: login}
	err := o.Read(&u)
	if err == orm.ErrNoRows {
		h := sha256.New()
		h.Write([]byte(pass))
		u.Pass = base64.StdEncoding.EncodeToString(h.Sum(nil))
		u.Name = u_name
		u.Login = login
		id, err := o.Insert(&u)
		if err != nil {
			Gologer.PError(err.Error())
			return 0,err
		}
		return id,nil
	}
	return 0, errors.New("user with this login already created")
}

func GetUserChats(user_id int64)([]*models.UserChatInfo, error){
	var final []*models.UserChatInfo
	type chatInfo struct{
		Id int64
		Name string
		Author_Id int64
		Type int
		Delete_Last int64
		Ban bool
	}
	type message struct{
		Chat_id int64
		LastSender string
		LastMessage string
		LastMessageTime int64
	}
	var meesagesBuffer[]message
	var ChatInfoBuffer []chatInfo
	qb, _ := orm.NewQueryBuilder(Driver)
	qb.Select("chats.id",
		"chats.name",
		"chats.author_id",
		"chats.type",
		"chat_users.delete_last",
		"chat_users.ban").
		From("chat_users").
		InnerJoin("chats").On("chat_users.chat_id = chats.id").
		Where("list_invisible = false").
		And("user_id = ?").
		Offset(0)
	sql := qb.String()
	o.Raw(sql, user_id).QueryRows(&ChatInfoBuffer)
	msg, _ := orm.NewQueryBuilder(Driver)
	msg.Select("messages.content",
		"messages.time",
		"users.name").
		From("messages").
		InnerJoin("users").
		On("messages.author_id = users.id").
		Where("messages.chat_id = ?").OrderBy("messages.time").Desc().Limit(1)
	sql = msg.String()
	for _,v:= range ChatInfoBuffer{
		o.Raw(sql, v.Id).QueryRows(&meesagesBuffer)
		var msg_now models.MessageContent
		err := json.Unmarshal([]byte(meesagesBuffer[0].LastMessage), msg_now)
		if err!=nil{
			Gologer.PError(err.Error())
			continue
		}
		var delete_v bool = true
		if v.Ban == false && v.Delete_Last == 0{
			delete_v = false
		}
		final = append(final,
			&models.UserChatInfo{
				ID:v.Id,
				Name: v.Name,
				Type: v.Type,
				LastSender: meesagesBuffer[0].LastSender,
				Admin_id: v.Author_Id,
				LastMessage: &msg_now,
				LastMessageTime: meesagesBuffer[0].LastMessageTime,
				View: 0,
				Delete: delete_v,
				Online: 0})
	}
	return final, nil
}

func GetUsersChatsIds(user_id int64)([]int64, error){
	var ids []int64
	qb, _ := orm.NewQueryBuilder(Driver)
	qb.Select("chat_id").
		From("chat_users").
		Where("user_id = ?")
	sql := qb.String()
	o.Raw(sql, user_id).QueryRows(&ids)
	return ids,nil
}

func GetOnlineUsersIdsInChats(chats_id*[]int64, users_online *[]int64)([]int64, error){
	var final []int64
	var users_in_strings []string
	for _,v := range *users_online{
		users_in_strings = append(users_in_strings, strconv.FormatInt(v, 10))
	}
	s:= strings.Join(users_in_strings, ",")
	//s= "("+s+")"

	var chats_in_string []string
	for _,v := range *chats_id{
		chats_in_string = append(chats_in_string, strconv.FormatInt(v, 10))
	}
	s1:= strings.Join(chats_in_string, ",")

	qb, err := orm.NewQueryBuilder(Driver)
	if err!= nil{
		return nil,err
	}
	qb.Select("user_id").
		From("chat_users").
		Where("user_id").In(s).
		And("ban = 0").
		And("list__invisible = 0").
		And("delete_last = 0").
		And("chat_id").In(s1)
	sql := qb.String()
	o.Raw(sql).QueryRows(&final)
	return final,nil
}

func GetUserSettings(user_id int64)(map[string]interface{}, error){
	var final = map[string]interface{}{}
	u:= User{Id: user_id}
	err := o.Read(&u)
	if err == orm.ErrNoRows {
		return final,errors.New("User not found")
	}
	final["login"] = u.Login
	final["name"]=u.Name

	return final, nil
}

func SetUserSettings(user_id int64, name string)(error){
	user := User{Id: user_id, Name: name }
	if _, err := o.Update(&user); err == nil {
		return err
	}
	return nil
}




