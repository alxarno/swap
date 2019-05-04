package db

import (
	"fmt"
	"log"

	"github.com/swap-messenger/swap/models"
	// "github.com/AlexeyArno/Gologer"
	"encoding/json"
	"strconv"
	strings "strings"
	"time"

	"github.com/astaxie/beego/orm"
)

const (
	//GetUserError - cannot get user by query
	GetUserError             = "Get user error: "
	//InsertChatError - insertion failed
	InsertChatError          = "Insert chat error: "
	InsertUserInChatError        = "Insert user in chat error: "
	InsertUserInChannelError     = "Insert user in channel error: "
	UserAlreadyExistInChatError = "User already in chat: "
	SendMessageError         = "Send message error: "
	GetChatError             = "Getting chat failed: "
	UserIsntAuthorError           = "User isnt author: "
	GetChatUserError        = "Gettings chat user failed: "
	GetDeletePointsError          = "Failed get delete points: "
	SetDeletePointsError          = "Failed set delete points: "
	UpdateChatUserErrro           = "Failed update chat user: "
	UpdateChatError                = "Failed update chat: "
	UserYetDidntDeleteError      = "User wasn't delete: "
)
// Create - creating chat or channel with the author's insertion
func Create(name string, authorID int64, chatChannel int)(int64, error){
	u := User{}

	db.First(&u, authorID)
	if u.ID == 0{
		return 0, DBE(GetUserError, nil)
	}
	tx := db.Begin()
	c := Chat{Name: name, Author: u, Type: chatChannel}
	if err := tx.Create(&c).Error; err != nil {
		tx.Rollback()
		return 0, DBE(InsertChatError, err)
	}
	tx.Commit()
	err := InsertUserInChat(u.ID, c.ID, false)
	if err != nil {
		if chatChannel == 0{
			return c.ID, DBE(InsertUserInChatError, err)
		}else{
			return c.ID, DBE(InsertUserInChannelError, err)
		}
	}
	if ChatCreated != nil {
		ChatCreated(authorID)
	}
	return c.ID, nil
}

// func CreateChat(name string, AuthorID int64) (int64, error) {
// 	u := User{}
// 	// err := db.QueryTable("users").Filter("id", AuthorId).
// 	// 	One(&u)
// 	db.First(&u, AuthorID)
// 	// if u.Id != 
// 	// err := db.First(&u, AuthorId)
// 	// if err != nil {
// 	// 	return 0, newError(GET_USER_ERROR + err.Error())
// 	// }
// 	tx := db.Begin()
// 	c := Chat{Name: name, Author: u, Type: 0}
// 	// tx.Create(&c)
// 	if err := tx.Create(&c).Error; err != nil {
// 		tx.Rollback()
// 		return 0, newError(INSERT_CHAT_ERROR + err.Error())
// 	}
// 	tx.Commit()
// 	err := InsertUserInChat(u.ID, c.ID, false)
// 	if err != nil {
// 		return c.ID, newError(INSERT_USER_IN_CHAT + err.Error())
// 	}
// 	// o.Commit()
// 	if ChatCreated != nil {
// 		ChatCreated(AuthorID)
// 	}
// 	return c.ID, nil
// }

// func CreateChannel(name string, AuthorID int64) (int64, error) {
// 	u := User{}
// 	db.First(&u, AuthorID)
// 	// err := o.QueryTable("users").Filter("id", AuthorId).One(&u)
// 	// if err != nil {
// 	// 	return 0, err
// 	// }
// 	c := Chat{Name: name, Author: &u, Type: 2}
// 	o.Begin()
// 	id, err := o.Insert(&c)
// 	if err != nil {
// 		o.Rollback()
// 		return 0, newError(INSERT_CHAT_ERROR + err.Error())
// 	}
// 	// o.Commit()

// 	err = InsertUserInChat(u.Id, id, false)
// 	if err != nil {
// 		o.Rollback()
// 		return id, newError(INSERT_USER_IN_CHANNEL + err.Error())
// 	}
// 	o.Commit()
// 	return id, nil
// }

func CheckUserInChatDelete(UserID int64, ChatID int64) (bool, error) {
	var cUser ChatUser
	// query := o.QueryTable("chat_users").Filter("user_id", UserId).Filter("chat_id", ChatId)
	// err := query.One(&cUser)
	db.Where("user_id = ?", UserID).Where("chat_id = ?", ChatID).Find(&cUser)

	if cUser.ID == 0 {
		return false, DBE(GetUserError,nil)
	}
	if cUser.ListInvisible || cUser.DeleteLast != 0 {
		return true, nil
	}
	return false, nil
}

func InsertUserInChat(UserId int64, ChatId int64, invite bool) error {
	var cUser ChatUser
	err := o.QueryTable("chat_users").Filter("user_id", UserId).Filter("chat_id", ChatId).One(&cUser)
	if err == nil {
		return newError(USER_ALREADY_EXIST_IN_CHAT)
	}

	cUser.User = &User{Id: UserId}
	cUser.Chat = &Chat{Id: ChatId}

	var DeletePoints [][]int64
	DeletePoints = append(DeletePoints, []int64{0, 0})
	cUser.Start = time.Now().Unix()
	cUser.SetDeletePoints(DeletePoints)
	o.Begin()
	_, err = o.Insert(&cUser)
	if err != nil {
		o.Rollback()
		return newError(INSERT_USER_IN_CHAT + err.Error())
	}
	o.Commit()
	var content string
	var command int
	if !invite {
		command = models.MESSAGE_COMMAND_USER_CREATED_CHAT
		// content = cUser.User.Name + " создал(а) беседу"
		if cUser.Chat.Type == 2 {
			command = models.MESSAGE_COMMAND_USER_CREATED_CHANNEL
			// content = cUser.User.Name + " создал(а) канал"
		}
	} else {
		command = models.MESSAGE_COMMAND_USER_INSERTED_TO_CHAT

		// content = cUser.User.Name + " приглашен(а) в беседу"
		if cUser.Chat.Type == 2 {
			command = models.MESSAGE_COMMAND_USER_INSERTED_TO_CHANNEL
			// content = cUser.User.Name + " приглашен(а) в канал"
		}
		if UserRequestedToChat != nil {
			UserRequestedToChat(UserId, ChatId, command)
		}
	}
	// log.Println("Command ", command)
	_, err = SendMessage(UserId, ChatId, content, 1, command)
	//UserRequestedToChat
	if err != nil {
		return newError(SEND_MESSAGE_ERROR + err.Error())
	}
	return nil
}

//?
func GetChatType(ChatId int64) (int, error) {
	var c Chat
	err := o.QueryTable("chat_users").Filter("id", ChatId).Filter("chat_id", ChatId).One(&c)
	if err != nil {
		return 0, err
	}
	return c.Type, nil
}

func CheckUserRightsInChat(UserId int64, ChatId int64) error {
	var c Chat

	err := o.QueryTable("chats").Filter("id", ChatId).One(&c)
	if err != nil {
		return newError(GET_CHAT_ERROR + err.Error())
	}

	// err := o.QueryTable("chat_users").Filter("id", ChatId).Filter("chat_id", ChatId).One(&c)
	// if err != nil {
	// 	return err
	// }
	if c.Author.Id != UserId {
		return newError(USER_ISNT_AUTHOR)
	}
	return nil
}

func GetChatsUsers(ChatId int64) ([]int64, error) {
	var users []int64
	qb, _ := orm.NewQueryBuilder(driver)

	qb.Select("user_id").
		From("chat_users").
		Where("chat_id = ?")

	sql := qb.String()

	o := orm.NewOrm()
	o.Raw(sql, ChatId).QueryRows(&users)
	return users, nil
}

func GetChatUserInfo(ChatId int64) (string, error) {
	type userInfo struct {
		ID         int    `json:"id"`
		Login      string `json:"login"`
		Name       string `json:"name"`
		DeleteLast int64  `json:"delete"`
		Ban        int    `json:"blocked"`
	}
	var data []userInfo
	qb, _ := orm.NewQueryBuilder(driver)

	qb.Select("users.id",
		"users.login",
		"users.name",
		"chat_users.delete_last",
		"chat_users.ban").
		From("chat_users").
		InnerJoin("users").On("users.id = chat_users.user_id").
		Where("chat_users.chat_id = ?").
		And("chat_users.list__invisible = 0")

	sql := qb.String()

	o.Raw(sql, ChatId).QueryRows(&data)
	for i, v := range data {
		if v.DeleteLast != 0 {
			data[i].DeleteLast = 1
		}
	}
	finish, _ := json.Marshal(data)
	return string(finish), nil
}

func DeleteUsersInChat(UserIDs []int64, ChatID int64, DeleteYourself bool) error {
	for _, v := range UserIDs {
		//c := ChatUser{User: &User{Id: v}, Chat:&Chat{Id: ChatId}, Delete_last: 0}
		//err:= o.Read(&c);if err!=nil{
		//	Gologer.PError(err.Error())
		//	continue
		//}
		var c ChatUser
		err := o.QueryTable("chat_users").Filter("user_id", v).
			Filter("chat_id", ChatID).Filter("delete_last", 0).One(&c)
		if err != nil {
			log.Println(GET_CHAT_USER_ERROR, err.Error(), map[string]interface{}{"UserID": v, "ChatID": ChatID})
			continue
		}
		dataPoints, err := c.GetDeletePoints()
		if err != nil {
			log.Println(GET_DELETE_POINTS, err.Error(), map[string]interface{}{"ChatUserID": c.Id})
			continue
		}
		if dataPoints[len(dataPoints)-1][0] == 0 {
			dataPoints[len(dataPoints)-1][0] = time.Now().Unix()
			c.Delete_last = dataPoints[len(dataPoints)-1][0]
			//fmt.Println(query)
			if DeleteYourself {
				c.Ban = false
			} else {
				c.Ban = true
			}
			err := c.SetDeletePoints(dataPoints)
			if err != nil {
				log.Println(SET_DELETE_POINTS, err.Error(),
					map[string]interface{}{"ChatUserID": c.Id, "deletePoints": dataPoints})
				continue
			}

		}
	}
	return nil
}

func RecoveryUsersInChat(UserIDs []int64, ChatID int64, RecoveryYourself bool) error {
	for _, v := range UserIDs {
		var c ChatUser
		err := o.QueryTable("chat_users").Filter("user_id", v).
			Filter("chat_id", ChatID).Filter("delete_last", 0).One(&c)
		if err != nil {
			log.Println(GET_CHAT_USER_ERROR, err.Error(), map[string]interface{}{"UserID": v, "ChatID": ChatID})
			continue
		}
		if RecoveryYourself {
			if c.Ban {
				continue
			}
		} else {
			c.Ban = false
		}

		deletePoints, err := c.GetDeletePoints()
		if err != nil {
			log.Println(GET_DELETE_POINTS, err.Error(), map[string]interface{}{"ChatUserID": c.Id})
			continue
		}
		if deletePoints[len(deletePoints)-1][1] == 0 {
			deletePoints[len(deletePoints)-1][1] = time.Now().Unix()
			deletePoints = append(deletePoints, []int64{0, 0})
			c.Delete_last = 0
			err := c.SetDeletePoints(deletePoints)
			if err != nil {
				log.Println(SET_DELETE_POINTS, err.Error(),
					map[string]interface{}{"ChatUserID": c.Id, "deletePoints": deletePoints})
				continue
			}
			_, err = o.Update(&c)
			if err != nil {
				log.Println(UPDATE_CHAT_USER, err.Error(), map[string]interface{}{"ChatUserID": c.Id})
				continue
			}
		}
	}
	return nil
}

func GetChatSettings(ChatId int64) (map[string]interface{}, error) {
	var sett = map[string]interface{}{}
	ch := Chat{Id: ChatId}
	err := o.Read(&ch)
	if err != nil {
		return sett, newError(GET_CHAT_ERROR + err.Error())
	}
	sett["name"] = ch.Name
	return sett, nil
}

func SetNameChat(ChatId int64, name string) error {
	ch := &Chat{Id: ChatId}
	err := o.Read(ch)
	if err != nil {
		return newError(GET_CHAT_ERROR + err.Error())
	}
	ch.Name = name
	_, err = o.Update(ch)
	if err != nil {
		return newError(UPDATE_CHAT + err.Error())
	}
	return nil
}

func DeleteChatFromList(UserId int64, ChatId int64) error {
	var c ChatUser
	err := o.QueryTable("chat_users").Filter("user_id", UserId).
		Filter("chat_id", ChatId).Filter("delete_last", 0).One(&c)
	if err != nil {
		return newError(GET_CHAT_USER_ERROR + err.Error())
	}
	res, err := CheckUserInChatDelete(UserId, ChatId)
	if err == nil && !res {
		return newError(USER_YET_DIDNT_DELETE)
	}
	c.List_Invisible = true
	_, err = o.Update(c)
	if err != nil {
		return newError(UPDATE_CHAT_USER + err.Error())
	}
	return nil
}

func FullDeleteChat(ChatId int64) error {
	var c Chat
	err := o.QueryTable("chats").Filter("id", ChatId).
		Filter("chat_id", ChatId).Filter("delete_last", 0).RelatedSel().One(&c)
	if err != nil {
		return newError(GET_CHAT_ERROR + err.Error())
	}
	var cu ChatUser
	err = o.QueryTable("chat_users").Filter("user_id", c.Author.Id).
		Filter("chat_id", ChatId).Filter("delete_last", 0).One(&cu)
	if err != nil {
		return newError(GET_CHAT_USER_ERROR + err.Error())
	}
	o.Delete(cu)
	qb, _ := orm.NewQueryBuilder(driver)

	qb.Delete().
		From("chat_users").
		Where("chat_id = ?")
	sql := qb.String()
	o.Raw(sql, ChatId).Exec()

	qb.Delete().
		From("messages").
		Where("chat_id = ?")
	sql = qb.String()
	o.Raw(sql, ChatId).Exec()
	//Need delete files
	o.Delete(&c)
	return nil
}

func GetUsersForAddByName(chatID int64, name string) ([]map[string]interface{}, error) {
	var findUsers []User
	var final []map[string]interface{}
	otherUsersIds, err := GetChatsUsers(chatID)
	if err != nil {
		return final, newError(GET_CHAT_USER_ERROR + err.Error())
	}

	var stringOtherUsersIds []string
	for _, v := range otherUsersIds {
		stringOtherUsersIds = append(stringOtherUsersIds, strconv.FormatInt(v, 10))
	}
	query := "SELECT id, name, login FROM users WHERE (id NOT IN (%s)) and ((name LIKE ?) or (login LIKE ?))"
	query = fmt.Sprintf(query, strings.Join(stringOtherUsersIds[:], ","))
	_, err = o.Raw(query, "%"+name+"%", "%"+name+"%").QueryRows(&findUsers)

	if err != nil {
		return final, newError(GET_USER_ERROR + err.Error())
	}

	for i, v := range findUsers {
		final = append(final, map[string]interface{}{})
		final[i]["name"] = v.Name
		final[i]["login"] = v.Login
		final[i]["id"] = v.Id
	}
	return final, nil
}
