package db

import (
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/swap-messenger/Backend/settings"
)

var (
	o      orm.Ormer
	driver = "mysql"
)

func LoadDb() {
	// register model

}

func createDB() error {
	err := orm.RunSyncdb("default", true, false)
	if err != nil {
		return err
	}
	o = orm.NewOrm()
	var sys System
	sys.Date = time.Now().Unix()
	sys.Version = "0.0.1"
	_, err = o.Insert(&sys)
	if err == nil {
		return err
	}
	return nil
}

func BeginDB() error {
	orm.RegisterDriver("sqlite3", orm.DRSqlite)
	sett, err := settings.GetSettings()
	if err != nil {
		panic(err)
	}
	if sett.Backend.Test {
		orm.Debug = true
		orm.RegisterDataBase("default", "sqlite3", "file:test.db")
	} else {
		orm.RegisterDataBase("default", "sqlite3", "file:"+settings.ServiceSettings.DB.SQLite.Path)
	}
	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Chat))
	orm.RegisterModel(new(ChatUser))
	orm.RegisterModel(new(Message))
	orm.RegisterModel(new(File))
	orm.RegisterModel(new(System))
	orm.RegisterModel(new(Dialog))

	o = orm.NewOrm()
	sys := System{}
	err = o.QueryTable("sys").Filter("id", 1).One(&sys)
	if err != nil {
		err = createDB()
		if err != nil {
			return err
		}
	}
	return nil
}
