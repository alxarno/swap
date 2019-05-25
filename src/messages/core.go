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

type systemMessage struct {
	data   string
	encode bool
}
type userConnection struct {
	UserID            int64
	MessageChan       chan models.NewMessageToUser
	EncryptedChan     chan models.EncryptedMessage
	SystemMessageChan chan string
	Auth              bool
	PublicKey         *rsa.PublicKey
}

type answer struct {
	MessageType string `json:"mtype"`
	Result      string `json:"result"`
	Action      string `json:"action"`
	Error       string `json:"error"`
	Key         string `json:"key,omniempty"`
}

var (
	debug    bool
	users    = []*userConnection{}
	entering = make(chan *userConnection)
	leaving  = make(chan *userConnection)

	sendMessages      = make(chan models.NewMessageToUser)
	forceSendMessages = make(chan models.ForceMsgToUser)
)

func (s *userConnection) writerUserSys(ws *websocket.Conn) {
	for sysMsg := range s.SystemMessageChan {

		if s.Auth {
			encryptedData, err := swapcrypto.EncryptMessage([]byte(sysMsg), s.PublicKey)
			if err != nil {
				if debug {
					log.Println("System encryption error - ", err.Error())
				}
				continue
			}
			encryptedMessage := models.EncryptedMessage{
				Data: encryptedData.Data, IV: encryptedData.IV, Key: encryptedData.Key,
				Type: messageEncrypted,
			}
			s.EncryptedChan <- encryptedMessage
			continue
		}

		if err := websocket.Message.Send(ws, string(sysMsg)); err != nil {
			if debug {
				log.Println(fmt.Sprintf("%s  %s", writingSystemChannelFailed, err.Error()))
			}
			break
		}
	}
}

func (s *userConnection) writerUser(ws *websocket.Conn) {
	for msg := range s.MessageChan {
		nowMessage, err := json.Marshal(msg)
		if err != nil {
			if debug {
				log.Println(fmt.Sprintf("Writer User -> %s  %s", marshalingMessageFailed, err.Error()))
			}
			continue
		}

		if s.Auth {
			encryptedData, err := swapcrypto.EncryptMessage(nowMessage, s.PublicKey)
			if err != nil {
				continue
			}
			encryptedMessage := models.EncryptedMessage{
				Data: encryptedData.Data, IV: encryptedData.IV, Key: encryptedData.Key,
				Type: messageEncrypted,
			}
			s.EncryptedChan <- encryptedMessage
			continue
		}

		if err := websocket.Message.Send(ws, string(nowMessage)); err != nil {
			if debug {
				log.Println(fmt.Sprintf("Writer User -> %s  %s", writingMessageChannelFailed, err.Error()))
			}
			break
		}
	}
}

func (s *userConnection) writeEncrypted(ws *websocket.Conn) {
	for emsg := range s.EncryptedChan {
		encryptedMessage, err := json.Marshal(emsg)
		if err != nil {
			if debug {
				log.Println(fmt.Sprintf("Writer Encrypted-> %s  %s", marshalingMessageFailed, err.Error()))
			}
			continue
		}
		if err := websocket.Message.Send(ws, string(encryptedMessage)); err != nil {
			if debug {
				log.Println(fmt.Sprintf("Writer Encrypted -> %s  %s", writingEncryptedChannelFailed, err.Error()))
			}
			break
		}
	}
}

func decodeNewMessage(msg string, connect *userConnection) {
	var data = make(map[string]interface{})
	var ans = answer{}

	messageTypeField := "mtype"
	// messageActionField := "action"
	// messagePayloadField := "payload"

	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		if debug {
			log.Println(fmt.Sprintf("Decode New Message -> %s  %s", unmarshalingMessageFailed, err.Error()))
		}
		return
	}
	switch data[messageTypeField] {
	case messageTypeSystem:
		systemM, err := systemMsg(msg)
		if err != nil {
			log.Println(fmt.Sprintf("Decode Sytem Message -> %s  %s", unmarshalingMessageFailed, err.Error()))
			return
		}

		if systemM.Action == messageActionAuth {
			token := systemM.Payload
			connect.PublicKey = swapcrypto.RsaPublicKeyByModulusAndExponent(
				systemM.N,
				systemM.E)
			if connect.PublicKey == nil {
				ans = answer{MessageType: messageTypeSystem, Result: messageFailed,
					Action: messageActionAuth, Error: "Public key is wrong"}
				if debug {
					log.Println("Public key is wrong")
				}
				finish, _ := json.Marshal(ans)
				connect.SystemMessageChan <- string(finish)
				return
			}
			user, err := api.TestUserToken(token)
			if err != nil {
				ans = answer{MessageType: messageTypeSystem, Result: messageFailed,
					Action: messageActionAuth, Error: err.Error()}
			} else {
				connect.UserID = user.ID
				connect.Auth = true
				ans = answer{MessageType: messageTypeSystem, Result: messageSuccess,
					Action: messageActionAuth, Key: string(swapcrypto.EncodedPublicKey)}
				userMove(connect.UserID, onlineUserInc)
			}
			finish, _ := json.Marshal(ans)
			connect.SystemMessageChan <- string(finish)
		}
		break
	case messageEncrypted:
		encryptedMessage, err := encryptedMsg(msg)
		if err != nil {
			log.Println(fmt.Sprintf("Decode Encrypted Message -> %s  %s", unmarshalingMessageFailed, err.Error()))
			return
		}
		// log.Println("Got encrypted message")
		msg, err := swapcrypto.DecryptMessage(encryptedMessage.Key, encryptedMessage.IV, encryptedMessage.Data)
		if err != nil {
			ans = answer{MessageType: messageTypeSystem, Result: messageSuccess,
				Action: messageActionAuth, Error: "Encryption is failed"}
			// log.Println("Decode Failed")
			finish, _ := json.Marshal(ans)
			connect.SystemMessageChan <- string(finish)
			return
		}
		// log.Println("Got encrypted message - ", msg)
		decodeNewMessage(msg, connect)
		return
	case messageTypeUser:
		log.Println(msg)
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
			// encryptedMessage, err := swapcrypto.EncryptMessage(finish, v.PublicKey)
			// if err != nil {
			// 	return err
			// }
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
			// encryptedMessage, err := swapcrypto.EncryptMessage(finish, v.PublicKey)
			// if err != nil {
			// 	return err
			// }
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
	user := &userConnection{}
	user.MessageChan = make(chan models.NewMessageToUser)
	user.SystemMessageChan = make(chan string)
	user.EncryptedChan = make(chan models.EncryptedMessage)
	go user.writeEncrypted(ws)
	go user.writerUser(ws)
	go user.writerUserSys(ws)

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
