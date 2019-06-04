package models

type MessageCommand int

const (
	MessageCommandNull                  MessageCommand = iota
	MessageCommandUserInsertedToChat                   // 1
	MessageCommandUserCreatedChat                      // 2
	MessageCommandUserInsertedToChannel                // 3
	MessageCommandUserCreatedChannel                   // 4
	MessageCommandUserInsertedToDialog                 // 5
	MessageCommandUserCreatedDialog                    // 6
	MessageCommandUserLeaveChat                        // 7
	MessageCommandUserReturnsToChat                    // 8
	MessageCommandUserWasBanned                        // 9
	MessageCommandUserWasUnbanned                      // 10
)

// const

//MessageType - type for message type's aliases
type MessageType int

const (
	//SystemMessageType - system message alias
	SystemMessageType MessageType = 1
	//UserMessageType - user message alias
	UserMessageType MessageType = 0
)

// const User
