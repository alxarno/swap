package db_api

import (
	"testing"
	"github.com/astaxie/beego/orm"
	"github.com/AlexeyArno/Gologer"
)
func init(){
	orm.RegisterDriver("sqlite3", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", "file:test.db")

	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Chat))
	orm.RegisterModel(new(Chat_User))
	orm.RegisterModel(new(Message))
	orm.RegisterModel(new(File))

	err := orm.RunSyncdb("default", true, false)
	if err != nil {
		//fmt.Println(err)
	}
	o = orm.NewOrm()
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
