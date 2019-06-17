package db2

import (
	"fmt"
	"testing"
)

const (
	testCannotCreateUser        = "User creating failed: "
	testCannotGetUsersForDialog = "Getting users for dialog failed: "
	testGotWrongUsersForDialog  = "Got wrong users for dialog creating: "
	testCreatingDialogFailed    = "Creating dialog failed: "
	testHavingDialogTestFailed  = "Checking users dialog failed: "
	testGettingDialogFailed     = "Getting dialog failed: "
	testGettingChatUserFailed   = "Getting chatuser failed: "
	testCreatedWrongChatUsers   = "Created chatusers with wrong chat: "
	testGettingChatFailed       = "Getting chat failed: "
)

func init() {
	createTestDB()
}

func TestGetUsersForCreateDialog(t *testing.T) {
	clearTestDB()
	//Creating 3 users
	user1 := User{Login: "user1", Pass: "1234"}
	user2 := User{Login: "user2", Pass: "1234"}
	user3 := User{Login: "user3", Pass: "1234"}
	var err error
	user1.ID, err = CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}
	user2.ID, err = CreateUser(user2.Login, user2.Pass, user2.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}
	user3.ID, err = CreateUser(user3.Login, user3.Pass, user3.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}
	//Get users for dialog for first user, should return 2 users (user2, user3)
	users, err := GetUsersForCreateDialog(user1.ID, "user")
	if err != nil {
		t.Error(testCannotGetUsersForDialog, err.Error())
		return
	}
	if len(*users) != 2 {
		t.Error(testGotWrongUsersForDialog,
			fmt.Sprintf("Users count should be 2 but got %d", len(*users)))
		return
	}
	//Creating dialog between user1 and user2
	_, err = CreateDialog(user1.ID, user2.ID)
	if err != nil {
		t.Error(testCreatingDialogFailed, err.Error())
		return
	}
	//Should return user3 because user1 and user2 already have dialog
	users, err = GetUsersForCreateDialog(user1.ID, "user")
	if err != nil {
		t.Error(testCannotGetUsersForDialog, err.Error())
		return
	}
	if len(*users) != 1 {
		t.Error(testGotWrongUsersForDialog,
			fmt.Sprintf("Users count should be 1 but got %d", len(*users)))
		return
	}
	if (*users)[0].ID != user3.ID {
		t.Error(testGotWrongUsersForDialog,
			fmt.Sprintf("User id should be %d but got %d", user3.ID, (*users)[0].ID))
		return
	}
}

func TestHaveAlreadyDialog(t *testing.T) {
	clearTestDB()
	//Creating 3 users
	user1 := User{Login: "user1", Pass: "1234"}
	user2 := User{Login: "user2", Pass: "1234"}
	user3 := User{Login: "user3", Pass: "1234"}
	var err error
	user1.ID, err = CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}
	user2.ID, err = CreateUser(user2.Login, user2.Pass, user2.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}
	user3.ID, err = CreateUser(user3.Login, user3.Pass, user3.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}
	_, err = HaveAlreadyDialog(user1.ID, user2.ID)
	if err == nil {
		t.Error(testHavingDialogTestFailed, "User1 and user2 yet dont have any dialog")
		return
	}
	//Creating dialog between user1 and user2
	_, err = CreateDialog(user1.ID, user2.ID)
	if err != nil {
		t.Error(testCreatingDialogFailed, err.Error())
		return
	}
	_, err = HaveAlreadyDialog(user1.ID, user2.ID)
	if err != nil {
		t.Error(testHavingDialogTestFailed, "User1 and user2 already have the dialog")
		return
	}
}

func TestCreateDialog(t *testing.T) {
	clearTestDB()
	//Creating 2 users
	user1 := User{Login: "user1", Pass: "1234"}
	user2 := User{Login: "user2", Pass: "1234"}
	var err error
	user1.ID, err = CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}
	user2.ID, err = CreateUser(user2.Login, user2.Pass, user2.Login)
	if err != nil {
		t.Error(testCannotCreateUser, err.Error())
		return
	}
	dialogID, err := CreateDialog(user1.ID, user2.ID)
	if err != nil {
		t.Error(testCreatingDialogFailed, err.Error())
		return
	}
	dialog := Dialog{ID: dialogID}
	if err = db.Where(&dialog).First(&dialog).Error; err != nil {
		t.Error(testGettingDialogFailed, err.Error())
		return
	}
	chatUser1 := ChatUser{UserID: dialog.User1ID}
	chatUser2 := ChatUser{UserID: dialog.User2ID}
	chat := Chat{ID: dialog.ChatID}
	if err = db.Where(&chatUser1).First(&chatUser1).Error; err != nil {
		t.Error(testGettingChatUserFailed, err.Error())
		return
	}
	if err = db.Where(&chatUser2).First(&chatUser2).Error; err != nil {
		t.Error(testGettingChatUserFailed, err.Error())
		return
	}
	if err = db.Where(&chat).First(&chat).Error; err != nil {
		t.Error(testGettingChatFailed, err.Error())
		return
	}

	if chatUser1.ChatID != dialog.ChatID ||
		chatUser2.ChatID != dialog.ChatID {
		t.Error(testCreatedWrongChatUsers, nil)
		return
	}

}
