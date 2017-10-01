package models

type Chat struct{
	ID float64
	Name string
	Addr_users []string
	MessageBlockId float64
	LastSender, LastMessage string
}
type MessageBlock struct {
	Chat_Id  float64
	Messages []Message
}
type Message struct {
	Addr_author string
	Content string
	Type string
	Chat_Id float64
}
type UserChatInfo struct{
	ID float64
	Name string
	//Addr_users []string
	LastSender string
	Admin_id float64
	Moders_ids []float64
	LastMessage *MessageContent
	LastMessageTime int64
	View int
	Delete int64
}
type MessageContent struct{
	Message *string
	Documents *[]string
	Type *string
}
type User struct {
	ID float64
	Name string
	Login string
	Pass string
}

type NewMessageToUser struct{
	Chat_Id *float64
	Content MessageContentToUser
	Author_id *float64
	Author_Name *string
	Time *int64
}

type MessageContentToUser struct{
	Message *string
	Documents []interface{}
	Type *string
}
type ForceMsgToUser struct{User_id float64; Msg NewMessageToUser}


func GetModels() string{
	return "Info"
}
