package db_api

import (
	"errors"
	//"github.com/astaxie/beego/orm"
	"github.com/AlexeyArno/Gologer"
	"time"
	"github.com/astaxie/beego/orm"
	"encoding/json"
)


func CreateChat(name string, AuthorId int64)(int64,error){
	u:= User{Id: AuthorId}
	err:= o.Read(&u);if err!=nil{
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
	u:= User{Id: AuthorId}
	err:= o.Read(&u);if err!=nil{
		return 0,err
	}
	c:= Chat{Name: name, Author: &u, Type: 2}
	id, err := o.Insert(&c);if err!=nil{
		return 0,err
	}
	return id,nil
}

func CheckUserInChatDelete(UserId int64, ChatId int64)(bool,error){
	UserInChat := chatUser{User:&User{Id: UserId}, Chat:&Chat{Id: ChatId}}
	err:= o.Read(&UserInChat);if err!=nil{
		return false,err
	}
	if UserInChat.List_Invisible|| UserInChat.Delete_last !=0{
		return true,nil
	}
	return false,nil
}

func InsertUserInChat(UserId int64, ChatId int64)(error){
	ChatUser:= chatUser{User:&User{Id:UserId}, Chat: &Chat{Id:ChatId}}
	err:=o.Read(&ChatUser);if err==nil{
		return errors.New("user already in chat")
	}
	DeletePoints := make([][]int64, 0)
	DeletePoints[0] = []int64{0,0}
	ChatUser.Start = time.Now().Unix()
	ChatUser.SetDeletePoints(DeletePoints)
	_,err = o.Insert(&ChatUser);if err!=nil{
		return err
	}
	return nil
}

func GetChatType(ChatId int64)(int,error){
	c:=Chat{Id: ChatId}
	err:=o.Read(&c);if err!=nil{
		return 0,err
	}
	return c.Type,nil
}

func CheckUserRightsInChat(UserId int64, ChatId int64)(error){
	c:= Chat{Id: ChatId}
	err:= o.Read(&c);if err!=nil{
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
		Where("chat_id = ?").
		Offset(0)

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
		ch_u:= chatUser{User: &User{Id: v}, Chat:&Chat{Id: ChatId}, Delete_last: 0}
		err:= o.Read(&ch_u);if err!=nil{
			Gologer.PError(err.Error())
			continue
		}
		data_points,err:= ch_u.GetDeletePoints();if err!=nil{
			Gologer.PError(err.Error()+" in user data :"+ch_u.Delete_points)
			continue
		}
		if data_points[len(data_points)-1][0]==0 {
			data_points[len(data_points)-1][0] = time.Now().Unix()
			ch_u.Delete_last = data_points[len(data_points)-1][0]
			//fmt.Println(query)
			if DeleteYourself{
				ch_u.Ban = false
			}else{
				ch_u.Ban = true
			}
			err := ch_u.SetDeletePoints(data_points);if err!=nil{
				Gologer.PError("fail set delete points: "+err.Error())
				continue
			}

		}
	}
	return nil
}

func RecoveryUsersInChat(UserIds []int64, ChatId int64, RecoveryYourself bool)(error){
	for _,v:= range UserIds{
		ch_u:= chatUser{User: &User{Id: v}, Chat:&Chat{Id: ChatId}}
		err:= o.Read(&ch_u);if err!=nil{
			Gologer.PError(err.Error())
			continue
		}
		if RecoveryYourself{
			if ch_u.Ban{
				continue
			}
		}else{
			ch_u.Ban = false
		}

		delete_points,err:= ch_u.GetDeletePoints();if err!=nil{
			Gologer.PError(err.Error()+" in user data :"+ch_u.Delete_points)
			continue
		}
		if delete_points[len(delete_points)-1][1]==0 {
			delete_points[len(delete_points)-1][1] = time.Now().Unix()
			delete_points = append(delete_points, []int64{0,0})
			ch_u.Delete_last = 0
			//fmt.Println(query)
			err := ch_u.SetDeletePoints(delete_points);if err!=nil{
				Gologer.PError("fail set delete points: "+err.Error())
				continue
			}
			_,err=o.Update(&ch_u);if err !=nil{
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
	chatUser := chatUser{User:&User{Id: UserId}, Chat: &Chat{Id: ChatId}}
	err:= o.Read(&chatUser);if err!=nil{
		Gologer.PError(err.Error())
		return err
	}
	res,err:= CheckUserInChatDelete(UserId,ChatId);if err==nil && !res{
		return errors.New("user yet not delete")
	}
	chatUser.List_Invisible = true
	_,err=o.Update(chatUser);if err!=nil{
		return err
	}
	return nil
}

func FullDeleteChat(ChatId int64)(error){
	ch:= Chat{Id: ChatId}
	err:= o.Read(&ch); if err!=nil{
		return err
	}
	ChatUser:= chatUser{Chat:&ch, User:&User{Id: ch.Author.Id}}
	err= o.Read(&ChatUser); if err!=nil{
		return err
	}
	o.Delete(ChatUser)
	qb, _ := orm.NewQueryBuilder(driver)
	//Delete users in caht
	qb.Delete().
		From("chat_users").
		Where("chat_id = ?").
		Offset(0)
	sql := qb.String()
	o.Raw(sql, ChatId).Exec()
	//Delete messages
	qb.Delete().
		From("messages").
		Where("chat_id = ?").
		Offset(0)
	sql = qb.String()
	o.Raw(sql, ChatId).Exec()
	//Need delete files

	o.Delete(&ch)
	return nil
}




