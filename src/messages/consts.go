package messageengine

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
	messageActionAuth             = "auth"
	messageSuccess                = "Success"
	messageFailed                 = "Error"
)

const (
	writingSystemChannelFailed  = "Writing to system channel failed: "
	writingMessageChannelFailed = "Writing message to channel failed: "
	marshalingMessageFailed     = "Message marshaling failed: "
	unmarshalingMessageFailed   = "Message unmarshaling failed: "
)

const ()
