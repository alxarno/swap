package db_api

import (
	"testing"
	"github.com/astaxie/beego/orm"
	"github.com/AlexeyArno/Gologer"
	"fmt"
)
func init(){
	o = orm.NewOrm()
}

func TestCreateUser(t *testing.T) {
	u:= User{Name: "Alex", Login: "alex123"}
	id, err:= CreateUser(u.Login, "1111", u.Name); if err!=nil{
		t.Error("Failed create user:", err)
	}
	u.Id = id
	err = o.Read(&u)
	if err == orm.ErrNoRows {
		t.Error("Failed found created user:")
	}
	o.Delete(&User{Id:id})
}

func TestGetOnlineUsersIdsInChats(t *testing.T) {
	var chats  []int64
	var users  []int64
	var cus []int64
	for i:=0;i<5;i++ {
		u := User{Name: "Alex", Login: "Hello"}
		id, err := o.Insert(&u)
		if err != nil {
			Gologer.PError("Failed insert")
			return
		}

		c := Chat{Name: "Hello", Author: &u, Type: 0}
		c_id, err := o.Insert(&c)
		if err != nil {
			Gologer.PError("Failed insert")
			return
		}

		cu := Chat_User{User: &u, Chat: &c}
		cu_id, err := o.Insert(&cu)
		if err != nil {
			Gologer.PError("Failed insert")
			return
		}
		chats = append(chats, c_id)
		users = append(users, id)
		cus = append(cus, cu_id)
	}
	final, err:= GetOnlineUsersIdsInChats(&chats, &users)
	if err != nil {
		t.Error("Returned error :",err.Error())
		return
	}
	if len(final) != len(users){
		t.Error("Answer length wrong")
		Gologer.PInfo(fmt.Sprintf("%d; need %d", len(final), len(users)))
		return
	}
	for i := range users {
		if final[i] != users[i] {
			t.Error("Answer wrong:","index :",i, "values:", final[i], users[i])
			return
		}
	}
	for _,v := range cus{
		o.Delete(&Chat_User{Id: v})
	}
	for _,v := range chats{
		o.Delete(&Chat{Id: v})
	}
	for _,v := range users{
		o.Delete(&User{Id: v})
	}
}

func TestGetUsersChatsIds(t *testing.T) {
	var chats  []int64
	var cus []int64
	u := User{Name: "Alex", Login: "Hello"}
	id, err := o.Insert(&u)
	if err != nil {
		Gologer.PError("Failed insert")
		return
	}
	for i:=0;i<5;i++ {
		c := Chat{Name: "Hello", Author: &u, Type: 0}
		c_id, err := o.Insert(&c)
		if err != nil {
			Gologer.PError("Failed insert")
			return
		}
		cu := Chat_User{User: &u, Chat: &c}
		cu_id, err := o.Insert(&cu)
		if err != nil {
			Gologer.PError("Failed insert")
			return
		}
		chats = append(chats, c_id)
		cus = append(cus, cu_id)
	}
	final, err:= GetUsersChatsIds(id)
	if err != nil {
		t.Error("Returned error :",err.Error())
		return
	}
	if len(final) != len(chats){
		t.Error("Answer length wrong")
		return
	}
	for i := range chats {
		if final[i] != chats[i] {
			t.Error("Answer wrong:","index :",i, "values:", final[i], chats[i])
			return
		}
	}
	for _,v := range cus{
		o.Delete(&Chat_User{Id: v})
	}
	for _,v := range chats{
		o.Delete(&Chat{Id: v})
	}
	o.Delete(&u)
}

func TestGetUserChats(t *testing.T) {

}

func TestGetUserSettings(t *testing.T) {
	u:= User{Login: "Alex111", Pass:"1111", Name:"Alex"}
	id,err := CreateUser(u.Login, u.Pass, u.Name);if err!=nil{
		t.Error(err)
	}
	data,err := GetUserSettings(id); if err!=nil{
		t.Error(err)
	}
	if data["login"].(string) != u.Login && data["name"] != u.Name{
		t.Error("Data is wrong")
	}
	o.Delete(&User{Id: id})
}

func TestSetUserSettings(t *testing.T) {
	newName := "jora"
	u:= User{Login: "Alex111", Pass:"1111", Name:"Alex"}
	id,err := CreateUser(u.Login, u.Pass, u.Name);if err!=nil{
		t.Error(err)
	}
	err = SetUserSettings(id, newName); if err!=nil{
		t.Error(err)
	}
	data,err := GetUserSettings(id); if err!=nil{
		t.Error(err)
	}
	if data["login"].(string) != u.Login && data["name"] != newName{
		t.Error("Data is wrong")
	}
	o.Delete(&User{Id: id})
}