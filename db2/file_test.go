package db2

import (
	"testing"
)

const (
	testCreatingFileFailed = "File creating failed: "
)

func TestCreateFile(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user1 := User{Login: "user1", Pass: "1234"}
	user2 := User{Login: "user2", Pass: "1111"}
	var err error
	user1.ID, err = CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user2.ID, err = CreateUser(user2.Login, user2.Pass, user2.Login)
	if err != nil {
		t.Error(testCannotCreateSecondUser, err.Error())
		return
	}
	chatName := "chat1"
	chatID, err := Create(chatName, user1.ID, ChatType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}

	_, _, err = CreateFile("file1.png", 1<<20, user1.ID, chatID, 1.9)
	if err != nil {
		t.Error(testCreatingFileFailed, err.Error())
		return
	}

	_, _, err = CreateFile("file2.png", 1<<20, user2.ID, chatID, 1.5)
	if err == nil {
		t.Error("User isn't in chat but send file to it")
		return
	}
	//
}

func TestDeleteFile(t *testing.T) {
	//
}

func TestCheckFileRights(t *testing.T) {
	//
}
