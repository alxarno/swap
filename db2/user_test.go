package db2

import (
	"fmt"
	"testing"
)

const (
	testCannotCreateFirstUser    = "Cannot create first user: "
	testCannotCreateSecondUser   = "Cannot create second user: "
	testWrongID                  = "Created record with wrong ID: "
	testCreatedUserWithSameLogin = "Created user with the same login: "
	testGotWrongUser             = "Got wrong user: "
	testGotWrongSettings         = "Got wrong settings: "
	testGotUserChats             = "Getting user chats failed: "
	testGotWornUserChats         = "Got wrong user chats: "
	testGetUsersChatsIDsError    = "Getting users chats IDs failed: "
	testGotWrongUserChatsIDs     = "Got wrong users chats IDs: "
)

func TestCreateUser(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user1 := User{Login: "user1", Pass: "1234"}
	user2 := User{Login: "user2", Pass: "1111"}

	// First user creation
	lindex, err := CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	// First record have 1 ID
	if lindex != 1 {
		t.Error(testWrongID)
		return
	}
	_, err = CreateUser(user1.Login, user1.Pass, user1.Login)
	if err == nil {
		t.Error(testCreatedUserWithSameLogin)
		return
	}

	lindex, err = CreateUser(user2.Login, user2.Pass, user2.Login)
	if err != nil {
		t.Error(testCannotCreateSecondUser, err.Error())
		return
	}
	// Second record have 2 ID
	if lindex != 2 {
		t.Error(testWrongID)
		return
	}

}

func TestGetUserByID(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(err.Error())
		return
	}
	u, err := GetUserByID(lindex)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err = db.Where("id = ?", lindex).First(&user).Error; err != nil {
		t.Error(err.Error())
		return
	}
	if u.ID != user.ID || u.Login != user.Login {
		t.Error(testGotWrongUser)
		return
	}
}

func TestGetUserByLoginAndPass(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(err.Error())
		return
	}
	u, err := GetUserByLoginAndPass(user.Login, user.Pass)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err = db.Where("id = ?", lindex).First(&user).Error; err != nil {
		t.Error(err.Error())
	}
	if u.ID != user.ID || u.Login != user.Login {
		t.Error(testGotWrongUser)
		return
	}
}

func TestGetUserChats(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	var err error
	user.ID, err = CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	_, err = Create("chat1", user.ID, DialogType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	chatsInfo, err := GetUserChats(user.ID)
	if err != nil {
		t.Error(testGotUserChats, err.Error())
		return
	}
	if len(*chatsInfo) != 1 {
		t.Error(testGotUserChats, fmt.Sprintf("Chats count should be equal 1, but got %d", len(*chatsInfo)))
		return
	}
	if (*chatsInfo)[0].AdminID != user.ID {
		t.Error(testGotWornUserChats,
			fmt.Sprintf("Chat's admin should be user with id %d, but got %d", user.ID, (*chatsInfo)[0].AdminID))
		return
	}
}

func TestGetUsersChatsIDs(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user.ID = lindex
	chatName := "chat1"
	chatID, err := Create(chatName, user.ID, DialogType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	chatsIDs, err := GetUsersChatsIDs(user.ID)
	if err != nil {
		t.Error(testGetUsersChatsIDsError, err.Error())
		return
	}
	if len(*chatsIDs) != 1 {
		t.Error(testGotWrongUserChatsIDs,
			fmt.Sprintf("Chat IDs should be 1, but got %d", len(*chatsIDs)))
		return
	}
	if (*chatsIDs)[0] != chatID {
		t.Error(testGotWrongUserChatsIDs,
			fmt.Sprintf("First chat ID should be 1, but got %d", (*chatsIDs)[0]))
		return
	}
}

func TestGetUserSettings(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(err.Error())
		return
	}
	settings, err := GetUserSettings(lindex)
	if err != nil {
		t.Error(err.Error())
		return
	}
	//Comparing name and login because we used login as name when we created user
	if settings.Name != user.Login {
		t.Error(testGotWrongSettings)
		return
	}
}

func TestSetUserSettings(t *testing.T) {
	createTestDB(t)
	defer deleteTestDB(t)
	newName := "user2"
	user := User{Login: "user1", Pass: "1234"}
	lindex, err := CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(err.Error())
		return
	}
	settings, err := GetUserSettings(lindex)
	if err != nil {
		t.Error(err.Error())
		return
	}
	settings.Name = newName
	if err = SetUserSettigns(lindex, settings); err != nil {
		t.Error(err.Error())
		return
	}
	settings, err = GetUserSettings(lindex)
	if err != nil {
		t.Error(err.Error())
		return
	}
	//Comparing name and login because we used login as name when we created user
	if settings.Name != newName {
		t.Error(testGotWrongSettings)
		return
	}
}
