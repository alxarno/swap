package db2

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/alxarno/swap/models"
	"github.com/alxarno/swap/settings"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type userRequestedCallback = func(userID int64, chatID int64, messageCommand models.MessageCommand)
type chatCreatedCallback = func(authorId int64)
type userDeleted = func(userID int64, chatID int64)
type sendUserMessage = func(mID int64, chatID int64, content *models.MessageContentToUser, authorID int64, time int64)

var (
	db         *gorm.DB
	testDBPath string
	//UserRequestedToChat - is a callback triggered when user requesting in chat(uses for sending notifications by WS)
	UserRequestedToChat userRequestedCallback
	//ChatCreated - is a callback triggered when user creating chat(uses for sending notifications by WS)
	ChatCreated chatCreatedCallback
	// SnedUserMessage - using for sending message
	SendUserMessageToSocket sendUserMessage
	// MessageAddedCallback
)

func createTestDB(t *testing.T) {
	var err error
	testDBPath = fmt.Sprintf("connection_%d.db", time.Now().UnixNano())
	db, err = gorm.Open("sqlite3", testDBPath)
	if err != nil {
		t.Error("Cannot create DB")
	}

	registerModels()
	db.LogMode(true)
}

func deleteTestDB(t *testing.T) {
	db.Close()
	if err := os.Remove(testDBPath); err != nil {
		t.Error("Cannot delete DB")
	}
}

func BeginDB() error {
	sett, err := settings.GetSettings()
	if err != nil {
		panic(err)
	}
	var dbPath = "test.db"
	if !sett.Backend.Test {
		dbPath = settings.ServiceSettings.DB.SQLite.Path
	}
	db, err = gorm.Open("sqlite3", dbPath)
	if err != nil {
		panic("Failed connect")
	}
	db.LogMode(true)

	registerModels()

	return nil
}

func registerModels() {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Chat{})
	db.AutoMigrate(&ChatUser{})
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&File{})
	db.AutoMigrate(&System{})
	db.AutoMigrate(&Dialog{})
}
