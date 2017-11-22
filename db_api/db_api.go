package db_api
import(
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
)
func init() {
	// register model


	// set default database
	orm.RegisterDriver("sqlite3", orm.DRSqlite)
	orm.RegisterDataBase("default", "sqlite3", "file:data.db")

	orm.RegisterModel(new(User))
	orm.RegisterModel(new(Chat))
	orm.RegisterModel(new(Chat_User))
	orm.RegisterModel(new(Message))
	orm.RegisterModel(new(File))

	err := orm.RunSyncdb("default", true, false)
	if err != nil {
		fmt.Println(err)
	}
}

func BeginDB(){

	o := orm.NewOrm()
	user := User{Name: "slene"}

	// insert
	_, err := o.Insert(&user)
	if err!=nil{
		fmt.Println(err)
	}

	// update
	user.Name = "astaxie"
	_, err = o.Update(&user)
	if err!=nil{
		fmt.Println(err)
	}
	// read one
	u := User{Id: user.Id}
	err = o.Read(&u)
	if err!=nil{
		fmt.Println(err)
	}

	// delete
	_,err = o.Delete(&u)
	if err!=nil{
		fmt.Println(err)
	}
}
