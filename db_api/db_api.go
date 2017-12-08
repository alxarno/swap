package db_api

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/Spatium-Messenger/Server/settings"
	//"github.com/AlexeyArno/Gologer"
)

var(
	o orm.Ormer
	driver = "mysql"
)
func init() {
	// register model
	orm.RegisterDriver("sqlite3", orm.DRSqlite)
	if settings.ServiceSettings.Server.Test {
		orm.RegisterDataBase("default", "sqlite3", "file:test.db")
	}else{
		orm.RegisterDataBase("default", "sqlite3", "file:data.db")
	}
	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Chat))
	orm.RegisterModel(new(chatUser))
	orm.RegisterModel(new(Message))
	orm.RegisterModel(new(File))
	orm.RegisterModel(new(Dialog))

	//err := orm.RunSyncdb("default", true, false)
	//if err != nil {
	//	//fmt.Println(err)
	//	Gologer.PError(err.Error())
	//}
	// set default database

}

func BeginDB(){
	//if !settings.Settings.Server.Test {
		o = orm.NewOrm()
	//}
}
