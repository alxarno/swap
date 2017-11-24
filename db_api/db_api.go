package db_api
import(
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
	//"fmt"
)
var(
	o orm.Ormer
)
//func init() {
//	// register model
//
//
//	// set default database
//	orm.RegisterDriver("sqlite3", orm.DRSqlite)
//	orm.RegisterDataBase("default", "sqlite3", "file:data.db")
//
//	orm.RegisterModel(new(User))
//	orm.RegisterModel(new(Chat))
//	orm.RegisterModel(new(Chat_User))
//	orm.RegisterModel(new(Message))
//	orm.RegisterModel(new(File))
//
//	err := orm.RunSyncdb("default", true, false)
//	if err != nil {
//		fmt.Println(err)
//	}
//}

func BeginDB(){
	//o = orm.NewOrm()
}
