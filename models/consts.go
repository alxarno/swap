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

// const User
