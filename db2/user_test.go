package db2

import (
	"testing"
)

const (
	testCannotCreateFirstUser    = "Cannot create first user: "
	testCannotCreateSecondUser   = "Cannot create second user: "
	testWrongID                  = "Created record with wrong ID: "
	testCreatedUserWithSameLogin = "Created user with the same login: "
	testGotWrongUser             = "Got wrong user: "
	testGotWrongSettings         = "Got wrong settings: "
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
