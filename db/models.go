package db

type User struct {
	ID       int64
	Login    string `orm:"size(32)"`
	Name     string `orm:"size(100)"`
	Pass     string `orm:"size(45)"`
	Chats    []ChatUser
	Messages []Message
	Files    []File
	MyChats  []Chat
	Dialogs  []Dialog
}

func (u *User) TableName() string {
	return "users"
}

type Chat struct {
	ID       int64  `orm:"auto"`
	Name     string `orm:"size(100)"`
	Author   User
	Type     int       `orm:"default(0)"`
	Files    []File    `orm:"reverse(many)"`
	Messages []Message `orm:"reverse(many)"`
}

func (u *Chat) TableName() string {
	return "chats"
}

type ChatUser struct {
	ID            int64 `orm:"auto"`
	User          User
	Chat          Chat
	Start         int64 `orm:"default(0)"`
	DeleteLast    int64 `orm:"default(0)"`
	DeletePoints  string
	Ban           bool `orm:"default(false)"`
	ListInvisible bool `orm:"default(false)"`
}

func (c *ChatUser) TableName() string {
	return "chat_users"
}

type Message struct {
	ID      int64 `orm:"auto"`
	Author  User
	Chat    Chat
	Content string
	Time    int64 `orm:"default(0)"`
}

func (u *Message) TableName() string {
	// db table name
	return "messages"
}

type File struct {
	ID        int64 `orm:"auto"`
	Author    User  `orm:"rel(fk)"`
	Chat      Chat  `orm:"rel(fk)"`
	Name      string
	Path      string
	RatioSize float64 `orm:"default(0)"`
	Size      int64   `orm:"default(0)"`
}

func (u *File) TableName() string {
	// db table name
	return "files"
}

type Dialog struct {
	ID     int64 `orm:"auto"`
	ChatID int64
	User1  User `orm:"rel(fk)"`
	User2  User `orm:"rel(fk)"`
}

func (u *Dialog) TableName() string {
	// db table name
	return "dialogs"
}

type System struct {
	ID      int64 `orm:"auto"`
	Date    int64 `orm:"default(0)"`
	Version string
}

func (u *System) TableName() string {
	// db table name
	return "sys"
}
