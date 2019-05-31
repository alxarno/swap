package db2

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alxarno/swap/models"
)

const (
	//DialogAlreadyCreated - dialog was already created
	DialogAlreadyCreated = "Dialog already created: "
	//CreatingChatFailed - chat creating failed
	CreatingChatFailed = "Chat creating failed: "
	//InsertingDialogFailed - inserting dialog in db failed
	InsertingDialogFailed = "Dialog inserting into db failed: "
	//GettingDialogFailed - dialog getting failed
	GettingDialogFailed = "Dialog getting failed: "
	//CheckingDialogExistsFailed - checking dialog exists failed
	CheckingDialogExistsFailed = "Checking dialog exists failed: "
	//UsersDontHaveDialog - users dont have dialog
	UsersDontHaveDialog = "Users doent have dialog: "
)

//GetUsersForCreateDialog - return pointer to users slice for probably starting dialog
//with supporting of name/login search
func GetUsersForCreateDialog(userID int64, name string) (*[]models.User, error) {
	response := []models.User{}
	userDialogs := []int64{}

	query := db.Model(&ChatUser{}).
		Joins("inner join chats on chats.id = chat_users.chat_id").
		Where("chats.type = ?", DialogType).
		Where("chat_users.user_id = ?", userID)

	if err := query.Pluck("chat_users.chat_id", &userDialogs).Error; err != nil {
		return nil, DBE(GetChatError, err)
	}
	name = "%" + name + "%"
	// GORM HAVE A BUG WITH BELOW 'IN" FUNCTION, doesnt processing []int...
	var UDialogsS []string
	for _, v := range userDialogs {
		UDialogsS = append(UDialogsS, strconv.FormatInt(v, 10))
	}
	s1 := strings.Join(UDialogsS, ",")
	//Get users info
	query = db.Model(&User{}).
		Select("users.id, users.name, users.login").
		// Where("chat_users.user_id = ?", userID).
		Joins("left join chat_users on chat_users.user_id = users.id").
		//Get users which not are already in dialog with currant user
		Where(fmt.Sprintf("(chat_users.chat_id not in (%s) or chat_users.user_id IS NULL)", s1)).
		Where("users.id <> ?", userID).
		// Where("chat_users.user_id <> ?", userID).
		Where("( users.name like ?", name).
		Or("users.login like ?)", name)
	// log.Println(query.Get("sql"))
	if err := query.Scan(&response).Error; err != nil {
		return nil, DBE(GetChatUserError, err)
	}
	return &response, nil
}

//HaveAlreadyDialog - return dialog ID if users have a dialog
func HaveAlreadyDialog(userID int64, yetuserID int64) (int64, error) {
	dialog := Dialog{ID: 0}
	query := db.Model(&Dialog{}).
		Where("user1_id = ? and user2_id = ?", userID, yetuserID).
		Or("user2_id = ? and user1_id = ?", userID, yetuserID)
	if query.First(&dialog).RecordNotFound() {
		return 0, DBE(UsersDontHaveDialog, nil)
	}
	return dialog.ID, nil
}

//CreateDialog - creating dialog between two users
func CreateDialog(userID int64, yetuserID int64) (int64, error) {
	dialogExists, err := HaveAlreadyDialog(userID, yetuserID)
	if err == nil {
		return 0, DBE(DialogAlreadyCreated, err)
	}
	if dialogExists != 0 {
		return 0, DBE(DialogAlreadyCreated, nil)
	}
	chatID, err := Create("", userID, DialogType)
	if err != nil {
		return 0, DBE(CreatingChatFailed, err)
	}
	err = InsertUserInChat(yetuserID, chatID, false)
	if err != nil {
		return 0, DBE(InsertUserInChatError, err)
	}
	dialog := Dialog{ChatID: chatID, User1ID: userID, User2ID: yetuserID}
	if err := db.Create(&dialog).Error; err != nil {
		return 0, DBE(InsertingDialogFailed, err)
	}
	return dialog.ID, nil
}
