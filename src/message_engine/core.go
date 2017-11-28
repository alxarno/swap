package message_engine
//User give us this {"type": "mes"/"system", "content": {
// 						{"chatId":2,"Content":{"Message":"...","Documents":["1","2"],"Type":"u_msg"},"Token":"eyJUeXA..."} /{"type":"authoriz", "Token": "asdasdw..."}
// 					}}
//We need parse this and choose what we should do with it
//
//We send user{"type": "mes"/"system", "content": {
// 						{"chatId":2,
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
	//methods "github.com/Spatium-Messenger/Server/src/api/methods"
	models "github.com/Spatium-Messenger/Server/models"
	"golang.org/x/net/websocket"
	"fmt"
	"encoding/json"
	//"github.com/Spatium-Messenger/Server/src/api/methods"
	"github.com/Spatium-Messenger/Server/db_api"
	"github.com/Spatium-Messenger/Server/src/api2"
)

type ConnectionSpatium struct {
	UserId int64
	MessChan chan models.NewMessageToUser
	SystemMessChan chan string
	Authoriz bool
}


type client chan<-models.NewMessageToUser
var(
	users 			  = []*ConnectionSpatium{}
	//online_users_ids = make(map[int64]interface{})
	//get_messages = make(chan string)  //get just json by user
	entering 		  = make(chan *ConnectionSpatium)
	leaving 		  = make(chan *ConnectionSpatium)

	sendMessages      = make(chan models.NewMessageToUser)
	forceSendMessages = make(chan models.ForceMsgToUser)

)

func UserMove(userId int64, mType string){
	userChats,err:= db_api.GetUsersChatsIds(userId)
	if err!=nil{
		return
	}
	var usersOnline []int64
	for _,b:=range users{
		if b.Authoriz==true{
			usersOnline = append(usersOnline, b.UserId)
		}
	}
	notificationIds,err:= db_api.GetOnlineUsersIdsInChats(&userChats, &usersOnline)
	if err!=nil{
		return
	}
	var data= make(map[string]interface{})
	data["action"] = "online_user"
	data["type"]= mType
	data["chats"]= userChats
	data["type_a"] = "system"
	data["self"] = false
	finish, _:=json.Marshal(data)
	for _,i := range notificationIds {
		for _,v :=range users{
			if v.UserId == i{
				if i== userId {
					if mType != "-" {
						data["self"] = true
						finish, _ := json.Marshal(data)
						v.SystemMessChan <- string(finish)
						data["self"] = false
					}
				}else{
					v.SystemMessChan<-string(finish)
				}
			}
		}
	}
}

func writerUserSys(ws *websocket.Conn,  sysCh<-chan string){
	for sysMsg := range sysCh{
		if err := websocket.Message.Send(ws, string(sysMsg)); err != nil {
			fmt.Println("Can't send")
			fmt.Println(err)
			break
		}
	}
}

func writerUser(ws *websocket.Conn, ch <-chan models.NewMessageToUser){
	for msg := range ch{
		nowMessage, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Fail Marshaling in function wruteUser :104")
			return
		}
		if err := websocket.Message.Send(ws, string(nowMessage)); err != nil {
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
			user, err:=api2.TestUserToken(token)
			var answer = make(map[string]interface{})
			if err!=nil{
				fmt.Println(err)
				answer["type_a"]="system"
				answer["result"]="Error"
				answer["action"]="authoriz"
				answer["type"]=err.Error()
			}else{
				connect.UserId = user.Id
				connect.Authoriz = true
				answer["type_a"]="system"
				answer["result"]="Success"
				answer["action"]="authoriz"
				UserMove(connect.UserId, "+")
			}
			finish, _:=json.Marshal(answer)
			connect.SystemMessChan<-string(finish)
			//fmt.Println(connect.UserId)
		}

	}else{
		messageToUser,err := UserMsg(msg)
		if err!=nil{
			fmt.Println(err.Error())
			return
		}
		sendMessages <-*messageToUser
	}
}


//Reload only chats list on client side
func SendNotificationAddUserInChat(userId int64)(error){
	var message = make(map[string]interface{})
	message["type_a"] = "system"
	message["action"] = "add_in_chat"
	finish, _:=json.Marshal(message)
	for _,v :=range users{
		if v.UserId == userId {
			v.SystemMessChan<-string(finish)
		}
	}
	return nil
}


//Reload  chats list and now chat window close on client side
func SendNotificationDeleteChat(userId int64)(error){
	var message = make(map[string]interface{})
	message["type_a"] = "system"
	message["action"] = "delete_chat"
	finish, _:=json.Marshal(message)
	for _,v :=range users{
		if v.UserId == userId{
			v.SystemMessChan<-string(finish)
		}
	}
	return nil
}



func GetOnlineUsersInChat(userId *[]int64)(int64){
	var count int64
	count= 0
	for _,v:= range users{
		for _,b := range *userId {
			if v.UserId == b{
				count+=1
			}
		}
	}
	return count
}

func SendMessage( msg models.NewMessageToUser){
	sendMessages <-msg
}

func SendForceMessage( msg models.ForceMsgToUser){
	forceSendMessages <-msg
}


func SocketListener(ws *websocket.Conn) {
	var err error
	ch:= make(chan  models.NewMessageToUser)
	systemChannel :=make(chan string)
	user:= &ConnectionSpatium{}
	user.MessChan = ch
	user.SystemMessChan = systemChannel
	go writerUser(ws, user.MessChan)
	go writerUserSys(ws, user.SystemMessChan)

	//go writerUser(ws, user.MessChan)
	entering<-user
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
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
		case msg:=<-sendMessages:
			chatsUsers,err:= db_api.GetChatsUsers(msg.ChatId)
			if err!=nil{
				fmt.Println(err)
				continue
			}
			for _,user  := range users{
				for _,v := range chatsUsers {
					if v == user.UserId{
						user.MessChan<-msg
					}
				}
			}
		case msg:=<-forceSendMessages:
			for _,user  := range users{
				if user.UserId == msg.UserId {
					user.MessChan<-msg.Msg
				}
			}
		case cli:=<-entering:
			users = append(users, cli)
		case cli:=<-leaving:
			//delete connection from list online users
			index := -1
			for i:=0;i<len(users);i++{
				if users[i] == cli{
					index=i
					UserMove(users[i].UserId, "-")
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