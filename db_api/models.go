package db_api


type User struct {
	Id   int64`orm:"auto"`
	Login string`orm:"size(32)"`
	Name string`orm:"size(100)"`
	Pass string`orm:"size(45)"`
	Chats []*Chat_User`orm:"reverse(many)"`
	Messages []*Message`orm:"reverse(many)"`
	Files []*File`orm:"reverse(many)"`
	My_Chats []*Chat`orm:"reverse(many)"`
}

func (u *User) TableName() string {
	return "users"
}

type Chat struct {
	Id   int64`orm:"auto"`
	Name string`orm:"size(100)"`
	Author *User`orm:"rel(fk)"`
	Type int`orm:"default(0)"`
	Files []*File`orm:"reverse(many)"`
	Messages []*Message`orm:"reverse(many)"`
}

func (u *Chat) TableName() string {
	return "chats"
}


type Chat_User struct {
	Id   int64`orm:"auto"`
	User *User`orm:"rel(fk)"`
	Chat *Chat`orm:"rel(fk)"`
	Start int`orm:"default(0)"`
	Delete_last int64`orm:"default(0)"`
	Delete_points string
	Ban bool`orm:"default(false)"`
	List_Invisible bool `orm:"default(false)"`
}

func (u *Chat_User) TableName() string {
	return "chat_users"
}


type Message struct {
	Id   int64`orm:"auto"`
	Author *User`orm:"rel(fk)"`
	Chat *Chat`orm:"rel(fk)"`
	Content string
	Time int64`orm:"default(0)"`
}

func (u *Message) TableName() string {
	// db table name
	return "messages"
}

type File struct {
	Id   int64`orm:"auto"`
	Author *User`orm:"rel(fk)"`
	Chat *Chat`orm:"rel(fk)"`
	Name string
	Path string
	Ratio_Size float64`orm:"default(0)"`
	Size int64`orm:"default(0)"`
}

func (u *File) TableName() string {
	// db table name
	return "files"
}
