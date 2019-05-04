package db2

import (
	"fmt"
	"testing"
)

const (
	testCreateChatError            = "Chat's creation is failed: "
	testChatCreatedWithWrongAuthor = "Chat created with wrong author ID: "
	testChatUserNotFound           = "Chat's user not found: "
	testGetChatTypeError           = "Getting chat's type failed: "
	testGotWrongChatType           = "Got wrong chat's type: "
	testCheckUserError             = "Got wrong user's rights check: "
	testGetChatUsersError          = "Getting chat users failed: "
	testGotWrongChatUsers          = "Got wrong chat's users: "
	testGetChatUsersInfo           = "Getting chat's users info failed: "
	testGotWrongChatUsersInfo      = "Got wrong chat's users info: "
	testDeleteUsersInChatFailed    = "Deleting users in chat failed: "
	testRecoveryUsersInChatFailed  = "Recovery users in chat failed: "
)

func TestCreate(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user1 := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	chatID, err := Create("chat1", lindex, ChatType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	_, err = Create("chat2", lindex+1, ChatType)
	if err == nil {
		t.Error(testChatCreatedWithWrongAuthor)
		return
	}
	chatUser := ChatUser{ChatID: chatID, UserID: lindex}
	if err = db.First(&chatUser).Error; err != nil {
		t.Error(testChatUserNotFound, err.Error())
		return
	}
}

func TestGetChatType(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user1 := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	chatID, err := Create("chat1", lindex, DialogType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	chattype, err := GetChatType(chatID)
	if err != nil {
		t.Error(testGetChatTypeError)
		return
	}
	if chattype != DialogType {
		t.Error(testGotWrongChatType, chattype)
		return
	}
}

func TestCheckUserRights(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user1 := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user1.ID = lindex
	user2 := User{Login: "user2", Pass: "1234"}
	lindex, err = CreateUser(user2.Login, user2.Pass, user2.Login)
	if err != nil {
		t.Error(testCannotCreateSecondUser, err.Error())
		return
	}
	user2.ID = lindex
	chatID, err := Create("chat1", user1.ID, DialogType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	if err = CheckUserRights(user1.ID, chatID); err != nil {
		t.Error(
			testCheckUserError,
			fmt.Sprintf("user %d have rights, but got negative", user1.ID),
		)
		return
	}
	if err = CheckUserRights(user2.ID, chatID); err == nil {
		t.Error(
			testCheckUserError,
			fmt.Sprintf("user %d haven't rights, but got positive", user2.ID),
		)
		return
	}
}

func TestGetChatsUsers(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user.ID = lindex
	chatID, err := Create("chat1", user.ID, DialogType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	users, err := GetChatsUsers(chatID)
	if err != nil {
		t.Error(testGetChatUsersError, err.Error())
		return
	}
	if len(*users) != 1 {
		t.Error(
			testGotWrongChatUsers,
			fmt.Sprintf("Chat's users count should equal 1 , but got %d", len(*users)),
		)
		return
	}
	if (*users)[0] != user.ID {
		t.Error(
			testGotWrongChatUsers,
			fmt.Sprintf("First chat's users id should equal 1 , but got %d", (*users)[0]),
		)
		return
	}
}

func TestGetChatUsersInfo(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user.ID = lindex
	chatID, err := Create("chat1", user.ID, DialogType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	usersInfo, err := GetChatUsersInfo(chatID)
	if err != nil {
		t.Error(testGetChatUsersInfo, err.Error())
		return
	}
	if len(*usersInfo) != 1 {
		t.Error(testGotWrongChatUsersInfo)
		return
	}
	if (*usersInfo)[0].ID != user.ID {
		t.Error(
			testGotWrongChatUsersInfo,
			fmt.Sprintf("First users info owned by user 1,but got %d", (*usersInfo)[0].ID),
		)
		return
	}
}

func TestDeleteUsersInChat(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user.ID = lindex
	chatID, err := Create("chat1", user.ID, DialogType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	err = DeleteUsersInChat([]int64{lindex}, chatID, true)
	if err != nil {
		t.Error(testDeleteUsersInChatFailed, err.Error())
		return
	}
	usersInfo, err := GetChatUsersInfo(chatID)
	if err != nil {
		t.Error(testGetChatUsersInfo, err.Error())
		return
	}
	if (*usersInfo)[0].DeleteLast != 1 {
		t.Error(testGotWrongChatUsersInfo,
			fmt.Sprintf("User 1 should be deleted but he don't, got %d", (*usersInfo)[0].DeleteLast))
		return
	}
}

func TestRecoveryUsersInChat(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user.ID = lindex
	chatID, err := Create("chat1", user.ID, DialogType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	err = DeleteUsersInChat([]int64{lindex}, chatID, true)
	if err != nil {
		t.Error(testDeleteUsersInChatFailed, err.Error())
		return
	}

	err = RecoveryUsersInChat([]int64{lindex}, chatID, true)
	if err != nil {
		t.Error(testRecoveryUsersInChatFailed, err.Error())
		return
	}
	usersInfo, err := GetChatUsersInfo(chatID)
	if err != nil {
		t.Error(testGetChatUsersInfo, err.Error())
		return
	}
	if (*usersInfo)[0].DeleteLast != 0 {
		t.Error(testGotWrongChatUsersInfo,
			fmt.Sprintf("User 1 should not be deleted but he does, got %d", (*usersInfo)[0].DeleteLast))
		return
	}
}
