package db2

import (
	"fmt"
	"testing"
)

const (
	testCreatingFileFailed     = "File creating failed: "
	testGettingFileFailed      = "Cannot get file: "
	testGotWrongFile           = "Got wrong file: "
	testDeletingFile           = "Deleting file failed: "
	testCannotInsertUserToChat = "Inserting user to chat was failed: "
	testGotWrongFileRights     = "Got wrong file rights: "
)

func init() {
	createTestDB()
}

func TestCreateFile(t *testing.T) {
	clearTestDB()
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

	fileID, filePath, err := CreateFile("file1.png", 1<<20, user1.ID, chatID, 1.9, 0)
	if err != nil {
		t.Error(testCreatingFileFailed, err.Error())
		return
	}

	_, _, err = CreateFile("file2.png", 1<<20, user2.ID, chatID, 1.5, 0)
	if err == nil {
		t.Error("User isn't in chat but send file to it")
		return
	}
	file := File{ID: fileID}
	if err = db.Where(&file).First(&file).Error; err != nil {
		t.Error(testGettingFileFailed, err.Error())
		return
	}
	if file.Path != filePath {
		t.Error(testGotWrongFile,
			fmt.Sprintf("File's path should be '%s' but got '%s'", filePath, file.Path))
		return
	}
	//
}

func TestDeleteFile(t *testing.T) {
	clearTestDB()
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
	chatID, err := Create("chat1", user1.ID, ChatType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	fileID, _, err := CreateFile("file1.png", 1<<20, user1.ID, chatID, 1.9, 0)
	if err != nil {
		t.Error(testCreatingFileFailed, err.Error())
		return
	}
	_, err = DeleteFile(fileID, user1.ID)
	if err != nil {
		t.Error(testDeletingFile, err.Error())
		return
	}

	file := File{ID: fileID}
	if err = db.Where(&file).First(&file).Error; err == nil {
		t.Error(testDeletingFile,
			fmt.Sprintf("File '%d' should be deleted", fileID))
		return
	}

	fileID, _, err = CreateFile("file2.png", 1<<20, user1.ID, chatID, 1.9, 0)
	if err != nil {
		t.Error(testCreatingFileFailed, err.Error())
		return
	}

	_, err = DeleteFile(fileID, user2.ID)
	if err == nil {
		t.Error(testDeletingFile, "File was deleted by no author")
		return
	}

}

func TestCheckFileRights(t *testing.T) {
	clearTestDB()
	user1 := User{Login: "user1", Pass: "1234"}
	user2 := User{Login: "user2", Pass: "1111"}
	user3 := User{Login: "user3", Pass: "1111"}
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
	user3.ID, err = CreateUser(user3.Login, user3.Pass, user3.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}

	chatID, err := Create("chat1", user1.ID, ChatType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	err = InsertUserInChat(user2.ID, chatID, true)
	if err != nil {
		t.Error(testCannotInsertUserToChat, err.Error())
	}
	fileID, _, err := CreateFile("file2.png", 1<<20, user1.ID, chatID, 1.9, 0)
	if err != nil {
		t.Error(testCreatingFileFailed, err.Error())
		return
	}

	_, err = CheckFileRights(user1.ID, fileID)
	if err != nil {
		t.Error(testGotWrongFileRights,
			"User1 should have rights because he is owner of chat and file creator")
		return
	}
	_, err = CheckFileRights(user2.ID, fileID)
	if err != nil {
		t.Error(testGotWrongFileRights,
			"User2 should have rights because he is chat member")
		return
	}
	_, err = CheckFileRights(user3.ID, fileID)
	if err == nil {
		t.Error(testGotWrongFileRights,
			"User3 shouldn't have rights because he isn't chat member")
		return
	}

	err = DeleteUsersInChat([]int64{user2.ID}, chatID, false)
	if err != nil {
		t.Error(testDeleteUsersInChatFailed, err.Error())
		return
	}
	_, err = CheckFileRights(user2.ID, fileID)
	if err == nil {
		t.Error(testGotWrongFileRights,
			"User2 shouldn't have rights because he was banned in chat")
		return
	}

}
