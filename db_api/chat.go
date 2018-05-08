package db_api

import (
	"errors"
	"github.com/AlexeyArno/Gologer"
	"time"
	"github.com/astaxie/beego/orm"
	"encoding/json"
	strings "strings"
	"strconv"
)


func CreateChat(name string, AuthorId int64)(int64,error){
	u:= User{}
	err:= o.QueryTable("users").Filter("id",AuthorId).
	One(&u);if err!=nil{
		return 0,err
	}
	c:= Chat{Name: name, Author: &u, Type: 0}
	id, err := o.Insert(&c);if err!=nil{
		return 0,err
	}
	err = InsertUserInChat(u.Id, id);if err!=nil{
		return id,err
	}
	return id,nil
}

func CreateChannel(name string, AuthorId int64)(int64,error){
	u:= User{}
	err:= o.QueryTable("users").Filter("id",AuthorId).One(&u)
	if err!=nil{
		return 0,err
	}
	c:= Chat{Name: name, Author: &u, Type: 2}
	id, err := o.Insert(&c);if err!=nil{
		return 0,err
	}
	return id,nil
}

func CheckUserInChatDelete(UserId int64, ChatId int64)(bool,error){
	//Gologer.PInfo(strconv.FormatInt(UserId,10))
	//Gologer.PInfo(strconv.FormatInt(ChatId,10))
	var cUser ChatUser
	err := o.QueryTable("chat_users").Filter("user_id", UserId).Filter("chat_id", ChatId).One(&cUser)
	if err!=nil{
		return false,err
	}
	if cUser.List_Invisible|| cUser.Delete_last !=0{
		return true,nil
	}
	return false,nil
}

func InsertUserInChat(UserId int64, ChatId int64)(error){
	var cUser ChatUser
	err := o.QueryTable("chat_users").Filter("user_id", UserId).Filter("chat_id", ChatId).One(&cUser)
	if err==nil{
		return errors.New("user already in chat")
	}

	cUser.User = &User{Id: UserId}
	cUser.Chat = &Chat{Id: ChatId}

	var DeletePoints [][]int64
	DeletePoints = append(DeletePoints, []int64{0,0})
	cUser.Start = time.Now().Unix()
	cUser.SetDeletePoints(DeletePoints)
	_,err = o.Insert(&cUser);if err!=nil{
		return err
	}
	content:= cUser.User.Name+" создал(а) беседу"
	if cUser.Chat.Type == 2{
		content= cUser.User.Name+" создал(а) канал"
	}
	_,err=SendMessage(ChatId,UserId,content,1);if err!=nil{
		Gologer.PError(err.Error())
		return err
	}
	return nil
}

func GetChatType(ChatId int64)(int,error){
	var c Chat
	err := o.QueryTable("chat_users").Filter("id", ChatId).Filter("chat_id", ChatId).One(&c)
	if err!=nil{
		return 0,err
	}
	return c.Type,nil
}

func CheckUserRightsInChat(UserId int64, ChatId int64)(error){
	var c Chat
	err := o.QueryTable("chat_users").Filter("id", ChatId).Filter("chat_id", ChatId).One(&c)
	if err!=nil{
		return err
	}
	if c.Author.Id != UserId{
		return errors.New("user haven't rights")
	}
	return nil
}

func GetChatsUsers(ChatId int64)([]int64,error){
	var users []int64
	qb, _ := orm.NewQueryBuilder(driver)

	qb.Select("user_id").
		From("chat_users").
		Where("chat_id = ?")

	sql := qb.String()

	o := orm.NewOrm()
	o.Raw(sql, ChatId).QueryRows(&users)
	return users,nil
}

func GetChatUserInfo(ChatId int64)(string,error){
	type userInfo struct {
		Id int `json:"id"`
		Login string `json:"login"`
		Name string `json:"name"`
		Delete_Last int64 `json:"delete"`
		Ban int `json:"blocked"`
	}
	var data []userInfo
	qb, _ := orm.NewQueryBuilder(driver)

	qb.Select("users.id",
		"users.login",
		"users.name",
		"chat_users.delete_last",
		"chat_users.ban").
		From("chat_users").
		InnerJoin("users").On("users.id = chat_users.user_id").
		Where("chat_users.chat_id = ?").
		And("chat_users.list__invisible = 0")

	sql := qb.String()

	o.Raw(sql, ChatId).QueryRows(&data)
	for i,v := range data{
		if v.Delete_Last!=0{
			data[i].Delete_Last = 1
		}
	}
	finish, _:=json.Marshal(data)
	return  string(finish), nil
}

func DeleteUsersInChat(UserIds []int64, ChatId int64, DeleteYourself bool)(error){
	for _,v:= range UserIds{
		//c := ChatUser{User: &User{Id: v}, Chat:&Chat{Id: ChatId}, Delete_last: 0}
		//err:= o.Read(&c);if err!=nil{
		//	Gologer.PError(err.Error())
		//	continue
		//}
		var c ChatUser
		err := o.QueryTable("chat_users").Filter("user_id", v).
		Filter("chat_id", ChatId).Filter("delete_last",0).One(&c)
		if err!=nil{
			Gologer.PError(err.Error())
			continue
		}
		dataPoints,err:= c.GetDeletePoints();if err!=nil{
			Gologer.PError(err.Error()+" in user data :"+ c.Delete_points)
			continue
		}
		if dataPoints[len(dataPoints)-1][0]==0 {
			dataPoints[len(dataPoints)-1][0] = time.Now().Unix()
			c.Delete_last = dataPoints[len(dataPoints)-1][0]
			//fmt.Println(query)
			if DeleteYourself{
				c.Ban = false
			}else{
				c.Ban = true
			}
			err := c.SetDeletePoints(dataPoints);if err!=nil{
				Gologer.PError("fail set delete points: "+err.Error())
				continue
			}

		}
	}
	return nil
}

func RecoveryUsersInChat(UserIds []int64, ChatId int64, RecoveryYourself bool)(error){
	for _,v:= range UserIds{
		var c ChatUser
		err := o.QueryTable("chat_users").Filter("user_id", v).
		Filter("chat_id", ChatId).Filter("delete_last",0).One(&c);if err!=nil{
			Gologer.PError(err.Error())
			continue
		}
		if RecoveryYourself{
			if c.Ban{
				continue
			}
		}else{
			c.Ban = false
		}

		deletePoints,err:= c.GetDeletePoints();if err!=nil{
			Gologer.PError(err.Error()+" in user data :"+ c.Delete_points)
			continue
		}
		if deletePoints[len(deletePoints)-1][1]==0 {
			deletePoints[len(deletePoints)-1][1] = time.Now().Unix()
			deletePoints = append(deletePoints, []int64{0,0})
			c.Delete_last = 0
			//fmt.Println(query)
			err := c.SetDeletePoints(deletePoints);if err!=nil{
				Gologer.PError("fail set delete points: "+err.Error())
				continue
			}
			_,err=o.Update(&c);if err !=nil{
				Gologer.PError("fail update user in chat info: "+err.Error())
				continue
			}
		}
	}
	return nil
}

func GetChatSettings(ChatId int64)(map[string]interface{}, error){
	var sett = map[string]interface{}{}
	ch:= Chat{Id: ChatId}
	err:=o.Read(&ch);if err!=nil{
		Gologer.PError(err.Error())
		return sett,err
	}
	sett["name"] = ch.Name
	return sett,nil
}

func SetNameChat(ChatId int64, name string)(error){
	ch:=Chat{Id: ChatId}
	err:= o.Read(ch);if err!=nil{
		Gologer.PError(err.Error())
		return err
	}
	ch.Name = name
	_,err = o.Update(ch); if err!=nil{
		Gologer.PError(err.Error())
		return err
	}
	return nil
}

func DeleteChatFromList(UserId int64, ChatId int64)(error){
	var c ChatUser
	err := o.QueryTable("chat_users").Filter("user_id",UserId).
	Filter("chat_id", ChatId).Filter("delete_last",0).One(&c);if err!=nil{
		Gologer.PError(err.Error())
		return err
	}
	res,err:= CheckUserInChatDelete(UserId,ChatId);if err==nil && !res{
		return errors.New("user yet not delete")
	}
	c.List_Invisible = true
	_,err=o.Update(c);if err!=nil{
		return err
	}
	return nil
}

func FullDeleteChat(ChatId int64)(error){
	var c Chat
	err := o.QueryTable("chats").Filter("id",ChatId).
	Filter("chat_id", ChatId).Filter("delete_last",0).RelatedSel().One(&c);if err!=nil{
		Gologer.PError(err.Error())
		return err
	}
	var cu ChatUser
	err = o.QueryTable("chat_users").Filter("user_id",c.Author.Id).
	Filter("chat_id", ChatId).Filter("delete_last",0).One(&cu);if err!=nil{
		Gologer.PError(err.Error())
		return err
	}
	o.Delete(cu)
	qb, _ := orm.NewQueryBuilder(driver)

	qb.Delete().
		From("chat_users").
		Where("chat_id = ?")
	sql := qb.String()
	o.Raw(sql, ChatId).Exec()

	qb.Delete().
		From("messages").
		Where("chat_id = ?")
	sql = qb.String()
	o.Raw(sql, ChatId).Exec()
	//Need delete files

	o.Delete(&c)
	return nil
}

func GetUsersForAddByName(chatId int64, name string)([]map[string]interface{},error){
	var findUsers []User
	var final []map[string]interface{}
	otherUsersIds,err:=GetChatsUsers(chatId);if err!=nil{
		Gologer.PError(err.Error())
		return final,err
	}

	var stringOtherUsersIds []string
	for _,v:=range otherUsersIds{
		stringOtherUsersIds = append(stringOtherUsersIds, strconv.FormatInt(v,10))
	}


	_,err = o.Raw("SELECT id, name,login FROM users WHERE id NOT IN (?) and ((name LIKE ?) or (login Like ?))",
		strings.Join(stringOtherUsersIds[:],","), "%"+name+"%","%"+name+"%").QueryRows(&findUsers)

	if err!=nil{
		Gologer.PError(err.Error())
		return final,err
	}
	for i,v:=range findUsers{
		final = append(final,map[string]interface{}{})
		final[i]["name"] = v.Name
		final[i]["login"] = v.Login
		final[i]["id"] = v.Id
	}
	return final,nil
}


