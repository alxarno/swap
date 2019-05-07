package db2

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/swap-messenger/swap/models"
	"github.com/swap-messenger/swap/settings"
)

type userRequestedCallback = func(userID int64, chatID int64, messageCommand models.MessageCommand)
type chatCreatedCallback = func(authorId int64)

var (
	db                  *gorm.DB
	testDBPath          string
	UserRequestedToChat userRequestedCallback = nil
	ChatCreated         chatCreatedCallback   = nil
)

func createTestDB(t *testing.T) {
	var err error
	testDBPath = fmt.Sprintf("connection_%d.db", time.Now().UnixNano())
	db, err = gorm.Open("sqlite3", testDBPath)
	if err != nil {
		t.Error("Cannot create DB")
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Chat{})
	db.AutoMigrate(&ChatUser{})
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&File{})
	db.AutoMigrate(&System{})
	db.AutoMigrate(&Dialog{})
	db.LogMode(true)
}

func deleteTestDB(t *testing.T) {
	db.Close()
	if err := os.Remove(testDBPath); err != nil {
		t.Error("Cannot delete DB")
	}
}

func BeginDB() error {

	// orm.RegisterDriver("sqlite3", orm.DRSqlite)
	sett, err := settings.GetSettings()
	if err != nil {
		panic(err)
	}
	var dbPath = "test.db"
	if !sett.Backend.Test {
		dbPath = settings.ServiceSettings.DB.SQLite.Path
	}
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		panic("Failed connect")
	}
	db.LogMode(false)

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Chat{})
	db.AutoMigrate(&ChatUser{})
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&File{})
	db.AutoMigrate(&System{})
	db.AutoMigrate(&Dialog{})

	// orm.RegisterModel(new(User))
	// orm.RegisterModel(new(Chat))
	// orm.RegisterModel(new(ChatUser))
	// orm.RegisterModel(new(Message))
	// orm.RegisterModel(new(File))
	// orm.RegisterModel(new(System))
	// orm.RegisterModel(new(Dialog))

	// o = orm.NewOrm()
	// sys := System{}
	// err = o.QueryTable("sys").Filter("id", 1).One(&sys)
	// if err != nil {
	// 	err = createDB()
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}
