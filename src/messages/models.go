package messageengine

import (
	"crypto/rsa"

	"github.com/swap-messenger/swap/models"
)

type onlineUsersMove int

const (
	onlineUserInc onlineUsersMove = 1
	onlineUserDec onlineUsersMove = 2
)

const (
	messageActionOnlineUser       = "online_user"
	messageActionUserChatInserted = "user_inserted"
	messageActionChatCreated      = "chat_created"
	messageActionUserAddedToChat  = "add_in_chat"
	messageActionDeleteChat       = "delete_chat"
	messageTypeSystem             = "system"
	messageTypeUser               = "user"
	messageEncrypted              = "encrypted"
	messageActionAuth             = "auth"
	messageActionKeyExchange      = "key-exchange"
	messageSuccess                = "Success"
	messageFailed                 = "Error"
)

const (
	writingSystemChannelFailed    = "Writing to system channel failed: "
	writingMessageChannelFailed   = "Writing message to channel failed: "
	marshalingMessageFailed       = "Message marshaling failed: "
	unmarshalingMessageFailed     = "Message unmarshaling failed: "
	writingEncryptedChannelFailed = "Writing to encrypted channel failed: "
)

const (
	messageTypeField = "mtype"
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
	KeyExchanged      bool
	PublicKey         *rsa.PublicKey
}

type answer struct {
	MessageType string `json:"mtype"`
	Result      string `json:"result"`
	Action      string `json:"action"`
	Error       string `json:"error"`
	Key         string `json:"key,omniempty"`
}
