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
	//db_work "github.com/AlexArno/spatium/db_work"
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
)

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
	fmt.Println(data)
	if data["type"] == "system"{
		action, err := SystemMsg(msg)
		if err!=nil{
			return
		}
		if action["Action"] == "Authoriz"{
			token := action["Payload"].(string)
			user, err:=methods.OnlyDecodeToken(secret, token)
				if err!=nil{
					fmt.Println(err)
					return
				}
			connect.UserId = user.ID
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


func SocketListener(ws *websocket.Conn) {
	var err error
	ch:= make(chan  models.NewMessageToUser)
	user:= &ConnectionSpatium{}
	user.MessChan = ch
	go writerUser(ws, user.MessChan)

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
			for _,user  := range users{
				user.MessChan<-msg
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