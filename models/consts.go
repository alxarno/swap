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
