package messageengine

import (
	"encoding/json"
	"log"

	swapcrypto "github.com/swap-messenger/swap/crypto"
	"github.com/swap-messenger/swap/models"
	"github.com/swap-messenger/swap/src/api"
)

func authHandler(msg SystemMessage, connect *userConnection) {
	var ans = answer{}
	token := msg.Payload
	user, err := api.TestUserToken(token)
	if err != nil {
		ans = answer{MessageType: messageTypeSystem, Result: messageFailed,
			Action: messageActionAuth, Error: err.Error()}
	} else {
		connect.UserID = user.ID
		connect.Auth = true
		ans = answer{MessageType: messageTypeSystem, Result: messageSuccess, Action: messageActionAuth}
		userMove(connect.UserID, onlineUserInc)
	}
	finish, _ := json.Marshal(ans)
	connect.SystemMessageChan <- string(finish)
}

func keyExchangeHandler(msg SystemMessage, connect *userConnection) {
	var ans = answer{}
	connect.PublicKey = swapcrypto.RsaPublicKeyByModulusAndExponent(
		msg.N,
		msg.E)
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
	connect.KeyExchanged = true
	ans = answer{MessageType: messageTypeSystem, Result: messageSuccess,
		Action: messageActionKeyExchange, Key: string(swapcrypto.EncodedPublicKey)}
	finish, _ := json.Marshal(ans)
	connect.SystemMessageChan <- string(finish)
}

func encryptedHandler(encryptedMessage models.EncryptedMessage, connect *userConnection) {
	var ans = answer{}
	decryptedMessage, err := swapcrypto.DecryptMessage(encryptedMessage.Key, encryptedMessage.IV, encryptedMessage.Data)
	if err != nil {
		ans = answer{MessageType: messageTypeSystem, Result: messageSuccess,
			Action: messageActionAuth, Error: "Encryption is failed"}
		finish, _ := json.Marshal(ans)
		connect.SystemMessageChan <- string(finish)
		return
	}
	decodeNewMessage(decryptedMessage, connect)
	return
}
