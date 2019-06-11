package db2

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/alxarno/swap/models"
)

const (
	testAddingMessageFailed  = "Message adding failed: "
	testGettingMessageFailed = "Getting message failed: "
	testMarshalingFailed     = "Marshaling failed: "
	testUnmarshalingFailed   = "Unmarshaling failed: "
	testGotWrongMessage      = "Got wrong message: "
	testSendingMessageFailed = "Sending message failed: "
)

func init() {
	createTestDB()
}

func TestAddMessage(t *testing.T) {
	// createTestDB()
	// defer deleteTestDB(t)
	clearTestDB()
	user1 := User{Login: "user1", Pass: "1234"}
	user2 := User{Login: "user2", Pass: "1234"}
	var err error
	lindex, err := CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user1.ID = lindex
	lindex, err = CreateUser(user2.Login, user2.Pass, user2.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user2.ID = lindex
	chatName := "chat1"
	chatID, err := Create(chatName, user1.ID, ChatType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	messageID, err := addMessage(user1.ID, chatID, "test")
	if err != nil {
		t.Error(testAddingMessageFailed, err.Error())
		return
	}
	message := Message{ID: messageID}
	if err = db.Where(&message).First(&message).Error; err != nil {
		t.Error(testGettingMessageFailed, err.Error())
		return
	}

	messageID, err = addMessage(user2.ID, chatID, "test")
	if err == nil {
		t.Error("User aren't in chat but has sent message to it")
		return
	}

	err = DeleteUsersInChat([]int64{user1.ID}, chatID, true)
	if err != nil {
		t.Error(testDeleteUsersInChatFailed, err)
		return
	}

	_, err = addMessage(user1.ID, chatID, "test")
	if err == nil {
		t.Error("User was deleted from chat but sent message to it")
		return
	}

}

func TestGetMessages(t *testing.T) {
	return
	clearTestDB()
	user := User{Login: "user1", Pass: "1234"}
	var err error
	user.ID, err = CreateUser(user.Login, user.Pass, user.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	chatID, err := Create("chat1", user.ID, ChatType)
	if err != nil {
		t.Error(testCreateChatError, err)
		return
	}

	content := models.MessageContent{
		Command: 0, Documents: []int64{}, Message: "hello", Type: int(models.UserMessageType)}

	jcontent, err := json.Marshal(content)
	if err != nil {
		t.Error(testMarshalingFailed, err.Error())
		return
	}
	c := string(jcontent)

	for i := 0; i < 100; i++ {
		_, err = addMessage(user.ID, chatID, c)
		if err != nil {
			t.Error(testAddingMessageFailed, err.Error())
			return
		}
	}

	//Getting first part message = 80 units
	messages, err := GetMessages(user.ID, chatID, false, 0)
	if err != nil {
		t.Error(testGettingMessageFailed, err)
		return
	}

	if len(*messages) != 80 {
		t.Errorf("Should be 80 messages, but got %d", len(*messages))
		return
	}
	//Getting first tranche (last 21 message)
	//(100 "added by user" + 1 "created within creating chat" - 80 "already recieved" = 21)
	messages, err = GetMessages(user.ID, chatID, true, (*messages)[len(*messages)-1].ID)
	if err != nil {
		t.Error(testGettingMessageFailed, err)
		return
	}
	if len(*messages) != 21 {
		t.Errorf("Should be 21 messages, but got %d", len(*messages))
		return
	}
	if (*(*messages)[len(*messages)-1].Content).Message != "hello" {
		t.Error(testGotWrongMessage,
			fmt.Sprintf("Last message content should be 'hello' but got '%s'",
				(*(*messages)[len(*messages)-1].Content).Message))
		return
	}
}

func TestSendMessage(t *testing.T) {
	clearTestDB()
	user1 := User{Login: "user1", Pass: "1234"}
	user2 := User{Login: "user2", Pass: "1234"}
	var err error
	lindex, err := CreateUser(user1.Login, user1.Pass, user1.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user1.ID = lindex
	lindex, err = CreateUser(user2.Login, user2.Pass, user2.Login)
	if err != nil {
		t.Error(testCannotCreateFirstUser, err.Error())
		return
	}
	user2.ID = lindex
	chatName := "chat1"
	chatID, err := Create(chatName, user1.ID, ChatType)
	if err != nil {
		t.Error(testCreateChatError, err.Error())
		return
	}
	messageID, err := AddMessage(user1.ID, chatID, "test", []int64{}, models.UserMessageType, models.MessageCommand(0))
	if err != nil {
		t.Error(testSendingMessageFailed, err.Error())
		return
	}

	message := Message{ID: messageID}
	if err = db.Where(&message).First(&message).Error; err != nil {
		t.Error(testGettingMessageFailed, err.Error())
		return
	}

	/********************************************************/
	content := models.MessageContent{}
	err = json.Unmarshal([]byte(message.Content), &content)
	if err != nil {
		t.Error(testUnmarshalingFailed, err.Error())
	}
	if content.Message != "test" {
		t.Error(testGotWrongMessage,
			fmt.Sprintf("Message's content should be 'test' but got '%s'", content.Message))
		return
	}
	/********************************************************/

	messageID, err = AddMessage(user2.ID, chatID, "test", []int64{}, models.UserMessageType, models.MessageCommandNull)
	if err == nil {
		t.Error("User aren't in chat but has sent message to it")
		return
	}

	err = DeleteUsersInChat([]int64{user1.ID}, chatID, true)
	if err != nil {
		t.Error(testDeleteUsersInChatFailed, err)
		return
	}

	messageID, err = AddMessage(user1.ID, chatID, "test", []int64{}, models.UserMessageType, models.MessageCommand(0))
	if err == nil {
		t.Error("User was deleted from chat but sent message to it")
		return
	}
}
