package models

type MessageCommand int

const (
	MessageCommandNull MessageCommand = iota
	MessageCommandUserInsertedToChat
	MessageCommandUserCreatedChat
	MessageCommandUserInsertedToChannel
	MessageCommandUserCreatedChannel
	MessageCommandUserInsertedToDialog
	MessageCommandUserCreatedDialog
)

const (
	MessageActionOnlineUser       = "online_sser"
	MessageActionUserChatInserted = "sser_inserted"
	MessageActionChatCreated      = "chat_created"
	MessageActionTypeSystem       = "system"
)
