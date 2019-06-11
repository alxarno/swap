package db2

type User struct {
	ID       int64
	Login    string     `gorm:"size:32"`
	Name     string     `gorm:"size:100"`
	Pass     string     `gorm:"size:45"`
	Language string     `gorm:"size:5;DEFAULT:en"`
	Chats    []ChatUser `gorm:"PRELOAD:false"`
	Messages []Message  `gorm:"PRELOAD:false"`
	Files    []File     `gorm:"PRELOAD:false"`
	MyChats  []Chat     `gorm:"PRELOAD:false"`
	Dialogs  []Dialog   `gorm:"PRELOAD:false"`
}

func (u *User) TableName() string {
	return "users"
}

type Chat struct {
	ID       int64
	Name     string `gorm:"size:100"`
	AuthorID int64
	Author   User
	Type     ChatMode `gorm:"DEFAULT:0"`
	Files    []File
	Messages []Message
}

func (u *Chat) TableName() string {
	return "chats"
}

type ChatUser struct {
	ID            int64
	UserID        int64
	User          User
	ChatID        int64
	Chat          Chat
	Start         int64  `gorm:"DEFAULT:0"`
	DeleteLast    int64  `gorm:"DEFAULT:0"`
	DeletePoints  string `gorm:"size:1024"`
	Ban           bool   `gorm:"DEFAULT:false"`
	ListInvisible bool   `gorm:"DEFAULT:false"`
}

func (c *ChatUser) TableName() string {
	return "chat_users"
}

type Message struct {
	ID       int64
	AuthorID int64
	Author   User
	ChatID   int64
	Chat     Chat
	Content  string `gorm:"size:1024"`
	Time     int64  `gorm:"DEFAULT:0"`
}

func (u *Message) TableName() string {
	// db table name
	return "messages"
}

type File struct {
	ID        int64
	AuthorID  int64
	Author    User
	ChatID    int64
	Chat      Chat
	Name      string
	Path      string
	RatioSize float64 `gorm:"DEFAULT:0"`
	Size      int64   `gorm:"DEFAULT:0"`
	Duration  int64   `gorm:"DEFAULT:0"`
}

func (u *File) TableName() string {
	// db table name
	return "files"
}

type Dialog struct {
	ID      int64
	ChatID  int64
	User1ID int64
	User1   User
	User2ID int64
	User2   User
}

func (u *Dialog) TableName() string {
	// db table name
	return "dialogs"
}

type System struct {
	ID      int64
	Date    int64 `gorm:"DEFAULT:0"`
	Version string
}

func (u *System) TableName() string {
	// db table name
	return "sys"
}

//Intermidiate types
type chatInfo struct {
	ID         int64
	Name       string
	AuthorID   int64
	Type       int
	DeleteLast int64
	Ban        bool
}
type message struct {
	name    string
	content string
	time    int64
}

type messageTemplate struct {
	ID       int64
	Content  string
	AuthorID int64
	Name     string
	Login    string
	Time     int64
}
