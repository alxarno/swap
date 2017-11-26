package db_api

import (
	"github.com/astaxie/beego/orm"
	"testing"
	"github.com/AlexeyArno/Gologer"
)

var u User

func init(){
	o = orm.NewOrm()
	u = User{Name:"Alex", Login:"Alex1111"}
	id,err := o.Insert(&u);if err!=nil{
		Gologer.PError(err.Error())
		return
	}
	u.Id = id

}

func TestCreateChat(t *testing.T) {
	id, err:= CreateChat("Pussy", u.Id); if err!=nil{
		t.Error(err.Error())
	}
	ch:= Chat{Id: id}
	err=o.Read(&ch);if err!=nil{
		t.Error(err.Error())
	}
	if ch.Name != "Pussy"{
		t.Error("Chat data is wrong")
	}
	o.Delete(ch)
}

func TestCreateChannel(t *testing.T) {
	id, err:= CreateChannel("Pussy", u.Id); if err!=nil{
		t.Error(err.Error())
	}
	ch:= Chat{Id: id}
	err=o.Read(&ch);if err!=nil{
		t.Error(err.Error())
	}
	if ch.Name != "Pussy" || ch.Type != 2{
		t.Error("Channel data is wrong")
	}
	o.Delete(&ch)
}

func TestInsertUserInChat(t *testing.T) {
	ch:= Chat{Name:"Hello", Author:&u}
	id,err := o.Insert(&ch);if err!=nil{
		t.Error(err.Error())
		return
	}
	u1:=User{Name:"alex2", Login:"Alex2"}
	UId,err := o.Insert(&u1);if err!=nil{
		t.Error(err.Error())
		return
	}
	err = InsertUserInChat(u.Id,id); if err==nil{
		t.Error("Double insert user")
	}
	err = InsertUserInChat(UId,id);if err!=nil{
		t.Error(err)
		return
	}
	ChatUser:= Chat_User{User: &User{Id:UId}, Chat:&Chat{Id:id}}
	err = o.Read(ChatUser); if err!=nil{
		t.Error(err)
		return
	}
	o.Delete(&ChatUser)
	o.Delete(&ch)
	o.Delete(&u1)
}

func TestCheckUserInChatDelete(t *testing.T) {
	ch:= Chat{Name:"Jiza"}
	id, err:= o.Insert(&ch);if err!=nil{
		t.Error(err)
		return
	}
	err = InsertUserInChat(u.Id, id); if err!=nil{
		t.Error(err)
		return
	}
	res,err:= CheckUserInChatDelete(u.Id, id); if err!=nil{
		t.Error(err)
	}
	if res{
		t.Error("wrong answer, user is undeleted")
	}
	ChatUser:= Chat_User{User:&u,Chat:&Chat{Id: ch.Id}}
	err=o.Read(&ChatUser);if err!=nil{
		t.Error(err)
	}
	ChatUser.Delete_last = 12312312
	_,err = o.Update(&ChatUser);if err!=nil{
		t.Error(err)
	}
	res,err = CheckUserInChatDelete(u.Id, id); if err!=nil{
		t.Error(err)
	}
	if !res{
		t.Error("wrong answer, user is undeleted")
	}
	ChatUser.Delete_last = 0
	ChatUser.List_Invisible = true
	_,err = o.Update(&ChatUser);if err!=nil{
		t.Error(err)
	}
	res,err = CheckUserInChatDelete(u.Id, id); if err!=nil{
		t.Error(err)
	}
	if !res{
		t.Error("wrong answer, user is undeleted")
	}
	ChatUser.Delete_last = 0
	ChatUser.List_Invisible = false
	_,err = o.Update(&ChatUser);if err!=nil{
		t.Error(err)
	}
	res,err = CheckUserInChatDelete(u.Id, id); if err!=nil{
		t.Error(err)
	}
	if res{
		t.Error("wrong answer, user is undeleted")
	}
	o.Delete(&ChatUser)
	o.Delete(&ch)
}

func TestGetChatType(t *testing.T) {
	id,err:=CreateChat("Alice", u.Id); if err!=nil{
		t.Error(err)
		return
	}
	CType,err:= GetChatType(id); if err!=nil{
		t.Error(err)
		return
	}
	if CType != 0{
		t.Error("wrong data")
		return
	}
	o.Delete(&Chat{Id: id})
}

func TestCheckUserRightsInChat(t *testing.T) {
	id, err:= CreateChat("Pussy", u.Id); if err!=nil{
		t.Error(err.Error())
	}
	err=CheckUserRightsInChat(u.Id, id); if err!=nil{
		t.Error("Wrong answer")
	}
	o.Delete(&Chat{Id:id})
}

func TestDeleteUsersInChat(t *testing.T) {
	ch:= Chat{Name:"Apollo", Author:&u}
	ch_id,err:= o.Insert(&ch); if err!=nil{
		t.Error(err)
		return
	}
	users:= []int64{}
	for i:=0;i<5;i++{
		u1:= User{Name:"hello"}
		id,err:= o.Insert(&u1); if err!=nil{
			t.Error(err)
			return
		}
		UserChat:= Chat_User{User:&u1, Chat:&Chat{Id: ch_id}}
		_,err = o.Insert(&UserChat); if err!=nil{
			t.Error(err)
			return
		}
		users= append(users, id)
	}

}


