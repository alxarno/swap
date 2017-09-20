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
	Addr_users []string
	LastSender, LastMessage string
	View int
}
type User struct {
	ID float64
	Name string
	Login string
	Pass string
}
func GetModels() string{
	return "Info"
}
