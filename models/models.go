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
	ID float64 `json:"id"`
	Name string `json:"name"`
	Type int64 `json:"type"`
	//Addr_users []string
	LastSender string `json:"last_sender"`
	Admin_id float64 `json:"admin_id"`
	Moders_ids []float64 `json:"moderators_ids"`
	LastMessage *MessageContent `json:"last_message"`
	LastMessageTime int64 `json:"last_message_time"`
	View int `json:"view"`
	Delete int64 `json:"delete"`
	Online int64 `json:"online"`
}
type MessageContent struct{
	Message *string `json:"content"`
	Documents *[]string `json:"documents"`
	Type *string `json:"type"`
}
type User struct {
	ID float64
	Name string
	Login string
	Pass string
}

type NewMessageToUser struct{
	ID *int64 `json:"id"`
	Chat_Id *float64 `json:"chat_id"`
	Content MessageContentToUser `json:"message"`
	Author_id *float64 `json:"author_id"`
	Author_Name *string `json:"author_name"`
	Author_Login *string `json:"author_login"`
	Time *int64 `json:"time"`
}

type MessageContentToUser struct{
	Message *string `json:"content"`
	Documents []interface{} `json:"documents"`
	Type *string `json:"type"`
}
type ForceMsgToUser struct{User_id float64; Msg NewMessageToUser}


func GetModels() string{
	return "Info"
}
