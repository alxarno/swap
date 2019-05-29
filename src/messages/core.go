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

	// "crypto/rsa"
	"encoding/json"
	"fmt"
	"log"

	models "github.com/alxarno/swap/models"

	// "github.com/alxarno/swap/src/api"
	"golang.org/x/net/websocket"

	db "github.com/alxarno/swap/db2"
)

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

		// if s.KeyExchanged && settings.ServiceSettings.Backend.Cert {
		// 	encryptedData, err := swapcrypto.EncryptMessage([]byte(sysMsg), s.PublicKey)
		// 	if err != nil {
		// 		if debug {
		// 			log.Println("System encryption error - ", err.Error())
		// 		}
		// 		continue
		// 	}
		// 	encryptedMessage := models.EncryptedMessage{
		// 		Data: encryptedData.Data, IV: encryptedData.IV, Key: encryptedData.Key,
		// 		Type: messageEncrypted,
		// 	}
		// 	s.EncryptedChan <- encryptedMessage
		// 	continue
		// }

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

		// if s.KeyExchanged && settings.ServiceSettings.Backend.Cert {
		// 	encryptedData, err := swapcrypto.EncryptMessage(nowMessage, s.PublicKey)
		// 	if err != nil {
		// 		continue
		// 	}
		// 	encryptedMessage := models.EncryptedMessage{
		// 		Data: encryptedData.Data, IV: encryptedData.IV, Key: encryptedData.Key,
		// 		Type: messageEncrypted,
		// 	}
		// 	s.EncryptedChan <- encryptedMessage
		// 	continue
		// }

		if err := websocket.Message.Send(ws, string(nowMessage)); err != nil {
			if debug {
				log.Println(fmt.Sprintf("Writer User -> %s  %s", writingMessageChannelFailed, err.Error()))
			}
			break
		}
	}
}

// func (s *userConnection) writeEncrypted(ws *websocket.Conn) {
// 	for emsg := range s.EncryptedChan {
// 		if !s.KeyExchanged || !settings.ServiceSettings.Backend.Cert {
// 			continue
// 		}
// 		encryptedMessage, err := json.Marshal(emsg)
// 		if err != nil {
// 			if debug {
// 				log.Println(fmt.Sprintf("Writer Encrypted-> %s  %s", marshalingMessageFailed, err.Error()))
// 			}
// 			continue
// 		}
// 		if err := websocket.Message.Send(ws, string(encryptedMessage)); err != nil {
// 			if debug {
// 				log.Println(fmt.Sprintf("Writer Encrypted -> %s  %s", writingEncryptedChannelFailed, err.Error()))
// 			}
// 			break
// 		}
// 	}
// }

func decodeNewMessage(msg string, connect *userConnection) {
	var data = make(map[string]interface{})
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		if debug {
			log.Println(fmt.Sprintf("Decode New Message -> %s  %s", unmarshalingMessageFailed, err.Error()))
		}
		return
	}
	switch data[messageTypeField] {
	case messageTypeSystem:
		sMessage, err := systemMsg(msg)
		if err != nil {
			log.Println(fmt.Sprintf("Decode Sytem Message -> %s  %s", unmarshalingMessageFailed, err.Error()))
			return
		}

		if sMessage.Action == messageActionKeyExchange {
			keyExchangeHandler(sMessage, connect)
		}

		if sMessage.Action == messageActionAuth {
			authHandler(sMessage, connect)
		}
		break
	case messageEncrypted:
		encryptedMessage, err := encryptedMsg(msg)
		if err != nil {
			log.Println(fmt.Sprintf("Decode Encrypted Message -> %s  %s", unmarshalingMessageFailed, err.Error()))
			return
		}
		encryptedHandler(encryptedMessage, connect)
		break
	case messageTypeUser:
		messageToUser, err := userMsg(msg)
		if err != nil {
			if debug {
				log.Println("Decode User Message -> " + err.Error())
			}
			return
		}
		sendMessages <- *messageToUser
	}
}

//ConnectionHandler - handles new WS connection
func ConnectionHandler(ws *websocket.Conn) {
	var err error
	user := &userConnection{}
	user.MessageChan = make(chan models.NewMessageToUser)
	user.SystemMessageChan = make(chan string)
	// user.EncryptedChan = make(chan models.EncryptedMessage)
	// go user.writeEncrypted(ws)
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