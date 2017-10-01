package message_engine
//User give us this {"type": "mes"/"system", "content": {
// 						{"Chat_Id":2,"Content":{"Message":"...","Documents":["1","2"],"Type":"u_msg"},"Token":"eyJUeXA..."} /{"type":"authoriz", "Token": "asdasdw..."}
// 					}}
//We need parse this and choose what we should do with it
//
//We send user{"type": "mes"/"system", "content": {
// 						{"Chat_Id":2,
// 						"Content":{"Message":"...","Documents":["1","2"],"Type":"u_msg"},
// 						"Author_Name": "Alex",
// 						"Author_ID": "1",
// 						"Time": 2132131231} / {"action": "notif", "Content":{}}
// 					}}
//
//
//
//
import(
	db_work "github.com/AlexArno/spatium/db_work"
	//methods "github.com/AlexArno/spatium/src/api/methods"
	models "github.com/AlexArno/spatium/models"
	"golang.org/x/net/websocket"
	"fmt"
	"encoding/json"
	//"github.com/AlexArno/spatium/src/api/methods"
	"github.com/AlexArno/spatium/src/api/methods"
)

type ConnectionSpatium struct {
	UserId float64
	MessChan chan models.NewMessageToUser
	SystemMessChan chan string
	Authoriz bool
}


type client chan<-models.NewMessageToUser
var(
	users = []*ConnectionSpatium{}
	//online_users_ids = make(map[int64]interface{})
	//get_messages = make(chan string)  //get just json by user
	entering = make(chan *ConnectionSpatium)
	leaving = make(chan *ConnectionSpatium)

	send_messages = make(chan models.NewMessageToUser)
	force_send_messages = make(chan models.ForceMsgToUser)
)

func writerUserSys(ws *websocket.Conn,  sys_ch<-chan string){
	for sysMsg := range sys_ch{
		if err := websocket.Message.Send(ws, string(sysMsg)); err != nil {
			fmt.Println("Can't send")
			fmt.Println(err)
			break
		}
	}
}

func writerUser(ws *websocket.Conn, ch <-chan models.NewMessageToUser){
	for msg := range ch{
		now_msg, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Fail Marshaling in function wruteUser :69")
			return
		}
		if err := websocket.Message.Send(ws, string(now_msg)); err != nil {
			fmt.Println("Can't send")
			fmt.Println(err)
			break
		}
	}
}


func decodeNewMessage(msg string, connect *ConnectionSpatium){
	var data= make(map[string]interface{})

	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		//panic(err)
		fmt.Println(err)
		return
	}
	//fmt.Println(data)
	if data["type"] == "system"{
		action, err := SystemMsg(msg)
		if err!=nil{
			return
		}
		if action["Action"] == "Authoriz"{
			token := action["Payload"].(string)
			user, err:=methods.OnlyDecodeToken(secret, token)
			var answer = make(map[string]interface{})
			if err!=nil{
				fmt.Println(err)
				answer["type_a"]="system"
				answer["result"]="Error"
				answer["action"]="authoriz"
				answer["type"]=err.Error()
			}else{
				connect.UserId = user.ID
				answer["type_a"]="system"
				answer["result"]="Success"
				answer["action"]="authoriz"
			}
			finish, _:=json.Marshal(answer)
			connect.SystemMessChan<-string(finish)
			//fmt.Println(connect.UserId)
		}

	}else{
		messageToUser,err := UserMsg(msg)
		if err!=nil{
			return
		}
		send_messages<-*messageToUser
	}
}

func SendNotificationAddUserInChat(user_id float64)(error){
	var message = make(map[string]interface{})
	message["type_a"] = "system"
	message["action"] = "add_in_chat"
	finish, _:=json.Marshal(message)
	for _,v :=range users{
		if v.UserId == user_id{
			v.SystemMessChan<-string(finish)
		}
	}
	return nil
}

func SendMessage( msg models.NewMessageToUser){
	send_messages<-msg
}

func SendForceMessage( msg models.ForceMsgToUser){
	force_send_messages<-msg
}


func SocketListener(ws *websocket.Conn) {
	var err error
	ch:= make(chan  models.NewMessageToUser)
	sysch:=make(chan string)
	user:= &ConnectionSpatium{}
	user.MessChan = ch
	user.SystemMessChan = sysch
	go writerUser(ws, user.MessChan)
	go writerUserSys(ws, user.SystemMessChan)

	//go writerUser(ws, user.MessChan)
	entering<-user
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive", err)
			break
		}
		decodeNewMessage(reply, user)

	}
	leaving<-user
	ws.Close()
}



func broadcaster(){
	for{
		select {
		case msg:=<-send_messages:
			chats_users,err:= db_work.GetChatsUsers(*msg.Chat_Id)
			if err!=nil{
				fmt.Println(err)
				continue
			}
			for _,user  := range users{
				for _,v := range chats_users{
					if v == user.UserId{
						user.MessChan<-msg
					}
				}
			}
		case msg:=<-force_send_messages:
			for _,user  := range users{
				if user.UserId == msg.User_id{
					user.MessChan<-msg.Msg
				}
			}
		case cli:=<-entering:
			users = append(users, cli)

			//clients[cli] = true
		case cli:=<-leaving:
			//delete connection from list online users
			index := -1
			for i:=0;i<len(users);i++{
				if users[i] == cli{
					index=i
				}
			}
			if index != -1{
				users[index] = users[len(users)-1]
				users=users[:len(users)-1]
			}
		}
	}
}

func StartCoreMessenger(){
	go broadcaster()
}