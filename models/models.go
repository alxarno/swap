package models

type Chat struct {
	ID                      int64
	Name                    string
	AddrUsers               []string
	MessageBlockID          int64
	LastSender, LastMessage string
}
type MessageBlock struct {
	ChatID   int64
	Messages []Message
}
type Message struct {
	AddrAuthor string
	Content    string
	Type       string
	ChatID     float64
}
type UserChatInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type int    `json:"type"`
	//Addr_users []string
	LastSender string `json:"last_sender"`
	AdminID    int64  `json:"admin_id"`
	//Moders_ids []float64 `json:"moderators_ids"`
	LastMessage     *MessageContent `json:"last_message"`
	LastMessageTime int64           `json:"last_message_time"`
	View            int             `json:"view"`
	Delete          bool            `json:"delete"`
	Online          int64           `json:"online"`
}
type MessageContent struct {
	Message   string  `json:"content"`
	Documents []int64 `json:"documents"`
	Type      string  `json:"type"`
	Command   int     `json:"command,integer"`
}
type User struct {
	ID    int64
	Name  string
	Login string
	Pass  string
}

type NewMessageToUser struct {
	ID          int64                 `json:"id"`
	ChatID      int64                 `json:"chat_id"`
	Content     *MessageContentToUser `json:"message"`
	AuthorID    int64                 `json:"author_id"`
	AuthorName  string                `json:"author_name"`
	AuthorLogin string                `json:"author_login"`
	Time        int64                 `json:"time"`
}

type CreateDHData struct {
	CommonName   string
	Organization string
	DNSNames1    string //wiki
	DNSNames2    string //192.168.0.2
	Type         string
}

type MessageContentToUser struct {
	Message   string                   `json:"content"`
	Documents []map[string]interface{} `json:"documents"`
	Type      string                   `json:"type"`
	Command   int                      `json:"command,integer"`
}

//type MessageContent struct{
//	Message string `json:"content"`
//	Documents []int64 `json:"documents"`
//	Type string `json:"type"`
//}

type ForceMsgToUser struct {
	UserID int64
	Msg    NewMessageToUser
}

func GetModels() string {
	return "Info"
}

type UserSettings struct {
	Name string
}

type ChatSettings struct {
	Name string `json:"name"`
}

type UserChatsInfo struct {
	ID         int64
	Login      string
	Name       string
	DeleteLast int64
	Ban        bool
}
