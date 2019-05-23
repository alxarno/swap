package messageengine

//
//			Client											MessageEngine											DB
//				|						Connect						|															|
//				|		---------------------->		|															|
//				|															|															|
//				|							Auth						|															|
//				|		---------------------->		|															|
//				|															|															|
//				|															|					Check User	    		|
//				|															|		--------------------->		|
//				|															|															|
//				|					Auth Result					|															|
//				|		<---------------------		|															|
//				|															|															|
//				|					Send Message    		|															|
//				|		--------------------->		|															|
//				|															|															|
//				|															|					Save Message    		|
//				|															|		--------------------->		|
//				|															|															|
//				|		Send Message To Users			|															|
//				|		<---------------------		|															|
//				|															|															|
//				|						Close							|															|
//				|		---------------------->		|															|
//				|															|															|
//

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"

	swapcrypto "github.com/swap-messenger/swap/crypto"
	models "github.com/swap-messenger/swap/models"
	"github.com/swap-messenger/swap/src/api"
	"golang.org/x/net/websocket"

	db "github.com/swap-messenger/swap/db2"
)

type userConnection struct {
	UserID            int64
	MessageChan       chan models.NewMessageToUser
	SystemMessageChan chan string
	Auth              bool
	PublicKey         *rsa.PublicKey
}

type answer struct {
	MessageType string            `json:"mtype"`
	Result      string            `json:"result"`
	Action      string            `json:"action"`
	Error       string            `json:"error"`
	Payload     interface{}       `json:"payload,omniempty"`
	Key         swapcrypto.JWKkey `json:"key"`
}

var (
	debug    bool
	users    = []*userConnection{}
	entering = make(chan *userConnection)
	leaving  = make(chan *userConnection)

	sendMessages      = make(chan models.NewMessageToUser)
	forceSendMessages = make(chan models.ForceMsgToUser)
)

func writerUserSys(ws *websocket.Conn, sysCh <-chan string) {
	for sysMsg := range sysCh {
		if err := websocket.Message.Send(ws, string(sysMsg)); err != nil {
			if debug {
				log.Println(fmt.Sprintf("%s  %s", writingSystemChannelFailed, err.Error()))
			}
			break
		}
	}
}

func writerUser(ws *websocket.Conn, ch <-chan models.NewMessageToUser) {
	for msg := range ch {
		nowMessage, err := json.Marshal(msg)
		if err != nil {
			if debug {
				log.Println(fmt.Sprintf("Writer User -> %s  %s", marshalingMessageFailed, err.Error()))
			}
			return
		}
		if err := websocket.Message.Send(ws, string(nowMessage)); err != nil {
			if debug {
				log.Println(fmt.Sprintf("Writer User -> %s  %s", writingMessageChannelFailed, err.Error()))
			}
			break
		}
	}
}

func decodeNewMessage(msg string, connect *userConnection) {
	var data = make(map[string]interface{})
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		if debug {
			log.Println(fmt.Sprintf("Decode New Message -> %s  %s", unmarshalingMessageFailed, err.Error()))
		}
		return
	}
	if data["type"] == messageTypeSystem {
		action, err := systemMsg(msg)
		if err != nil {
			return
		}
		if action["Action"] == messageActionAuth {
			var answer = answer{}
			token := action["Payload"].(string)
			connect.PublicKey = swapcrypto.RsaPublicKeyByModulusAndExponent(action["n"].(string), action["e"].(string))
			if connect.PublicKey == nil {
				answer.MessageType = messageTypeSystem
				answer.Result = messageFailed
				answer.Action = messageActionAuth
				answer.Error = "Public key is wrong"
				finish, _ := json.Marshal(answer)
				connect.SystemMessageChan <- string(finish)
				return
			}
			user, err := api.TestUserToken(token)
			if err != nil {
				answer.MessageType = messageTypeSystem
				answer.Result = messageFailed
				answer.Action = messageActionAuth
				answer.Error = err.Error()
			} else {
				connect.UserID = user.ID
				connect.Auth = true
				answer.MessageType = messageTypeSystem
				answer.Result = messageSuccess
				answer.Action = messageActionAuth
				answer.Key = swapcrypto.JWKPublicKey
				userMove(connect.UserID, onlineUserInc)
			}

			finish, _ := json.Marshal(answer)
			connect.SystemMessageChan <- string(finish)
		}

	} else {
		messageToUser, err := userMsg(msg)
		if err != nil {
			if debug {
				log.Println("decodeNewMessage error: " + err.Error())
			}
			return
		}
		sendMessages <- *messageToUser
	}
}

//SendNotificationAddUserInChat - Reload only chats list on client side
func SendNotificationAddUserInChat(userID int64) error {
	var message = answer{
		MessageType: messageTypeSystem,
		Action:      messageActionUserAddedToChat,
	}
	finish, _ := json.Marshal(message)
	for _, v := range users {
		if v.UserID == userID {
			v.SystemMessageChan <- string(finish)
		}
	}
	return nil
}

//SendNotificationDeleteChat - Reload  chats list and now chat window close on client side
func SendNotificationDeleteChat(userID int64) error {
	var message = answer{
		MessageType: messageTypeSystem,
		Action:      messageActionDeleteChat,
	}
	finish, _ := json.Marshal(message)
	for _, v := range users {
		if v.UserID == userID {
			v.SystemMessageChan <- string(finish)
		}
	}
	return nil
}

//GetOnlineUsersInChat - return count online users in certain chat
func GetOnlineUsersInChat(userIDs *[]int64) int64 {
	var count int64
	count = 0
	for _, v := range users {
		for _, b := range *userIDs {
			if v.UserID == b {
				count++
			}
		}
	}
	return count
}

//SendMessage - send message
func SendMessage(msg models.NewMessageToUser) {
	sendMessages <- msg
}

//SendForceMessage - send force message
func SendForceMessage(msg models.ForceMsgToUser) {
	forceSendMessages <- msg
}

//ConnectionHandler - handles new WS connection
func ConnectionHandler(ws *websocket.Conn) {
	var err error
	ch := make(chan models.NewMessageToUser)
	systemChannel := make(chan string)
	user := &userConnection{}
	user.MessageChan = ch
	user.SystemMessageChan = systemChannel
	go writerUser(ws, user.MessageChan)
	go writerUserSys(ws, user.SystemMessageChan)

	entering <- user
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			break
		}
		decodeNewMessage(reply, user)

	}
	leaving <- user
	ws.Close()
}

func broadcaster() {
	for {
		select {
		case msg := <-sendMessages:
			chatsUsers, err := db.GetChatsUsers(msg.ChatID)
			if err != nil {
				continue
			}
			for _, user := range users {
				for _, v := range *chatsUsers {
					if v == user.UserID {
						user.MessageChan <- msg
					}
				}
			}
		case msg := <-forceSendMessages:
			for _, user := range users {
				if user.UserID == msg.UserID {
					user.MessageChan <- msg.Msg
				}
			}
		case cli := <-entering:
			users = append(users, cli)
		case cli := <-leaving:
			//delete connection from list online users
			index := -1
			for i := 0; i < len(users); i++ {
				if users[i] == cli {
					index = i
					userMove(users[i].UserID, onlineUserDec)
				}
			}
			if index != -1 {
				users[index] = users[len(users)-1]
				users = users[:len(users)-1]
			}
		}
	}
}

//StartCoreMessenger - starting message engine
func StartCoreMessenger(_debug bool) {
	debug = _debug
	go broadcaster()
}
