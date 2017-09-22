package main

import (
	"fmt"
	"net/http"
	"log"
	"golang.org/x/net/websocket"
	"encoding/json"
	"strconv"
	db_work "github.com/AlexArno/spatium/db_work"
	models "github.com/AlexArno/spatium/models"
	api "github.com/AlexArno/spatium/src/api"
	"github.com/gorilla/mux"
)
//type Chat struct{
//	ID float64
//	Name string
//	Addr_users []string
//	MessageBlockId float64
//	LastSender, LastMessage string
//}
//type MessageBlock struct {
//	Chat_Id  float64
//	Messages []Message
//}
//type Message struct {
//	Addr_author string
//	Content string
//	Type string
//	Chat_Id float64
//}

// struct for user talk
//type UserChatInfo struct{
//	ID float64
//	Name string
//	Addr_users []string
//	LastSender, LastMessage string
//	View int
//}
type ProveConnection struct{
	Login string
	Pass string
}
type RequestGetMessage struct{
	Author string
	Chat_Id float64
}
type ErrorAnswer struct{
	Result string
	Type string
}


type client chan<-models.Message
var (
	chats []*models.Chat
	messagesBlock []*models.MessageBlock
	messages = make(chan models.Message)
	entering = make(chan client)
	leaving = make(chan client)
)


func broadcaster(){
	clients:= make(map[client]bool)
	for{
		select {
			case msg:=<-messages:
				for cli:= range clients{
					cli<-msg
				}
			case cli:=<-entering:
				clients[cli] = true
			case cli:=<-leaving:
				delete(clients, cli)
				close(cli)
		}
	}
}


func writerUser(ws *websocket.Conn, ch<-chan models.Message){
	for msg:=range ch{
		now_msg, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Fail Marshaling in function wruteUser :69")
			return
		}
		if err := websocket.Message.Send(ws, string(now_msg)); err != nil {
			fmt.Println("Can't send")
			break
		}
	}

}


func SocketListener(ws *websocket.Conn) {
	var err error
	ch:= make(chan  models.Message)
	go writerUser(ws, ch)
	entering<-ch
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive")
			break
		}

		//parse user request message and send them to saver
		byt := []byte(reply)
		var dat map[string]interface{}
		if err := json.Unmarshal(byt, &dat);err != nil{
			panic(err)
		}
		chat_id := dat["chat_id"].(float64)
		now_msg := models.Message{dat["Addr_author"].(string), dat["Content"].(string), dat["Type"].(string), chat_id}
		for v,r := range messagesBlock{
			if float64(r.Chat_Id) == chat_id{
				messagesBlock[v].Messages = append(messagesBlock[v].Messages, now_msg)
			}
		}
		for v,r := range chats{
			if float64(r.ID) == chat_id{
				chats[v].LastSender = now_msg.Addr_author
				chats[v].LastMessage = now_msg.Content
			}
		}
		//fmt.Println(chat_id)
		//fmt.Println("Received back from client: " + reply)
		messages<-now_msg
	}
	leaving<-ch
	ws.Close()
}

func createMainChat(id float64){
	//msg:=json.Marshal([]map{"type":"a_msg", "content": "God: i'm create this"})
	//msg:=`{"type":"a_msg", "content": "God: i'm create this"}`
	messageBloc1:= models.MessageBlock{id,[]models.Message{{"127.0.0.1:1234", "God: Im create this chat "+strconv.Itoa(int(id)), "a_msg", id}}}
	chat1 := models.Chat{id,strconv.Itoa(int(id)),[]string{"127.0.0.1:1234"}, id, "God", "i'm create chat"}
	messagesBlock = append(messagesBlock, &messageBloc1)
	chats = append(chats, &chat1)
}

func getChats(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var usersChats []*models.UserChatInfo
	for _,r := range  chats{
		usersChats = append(usersChats, &models.UserChatInfo{r.ID,r.Name,r.Addr_users,r.LastSender, r.LastMessage,0})
	}
	data, _ := json.Marshal(usersChats)
	fmt.Fprintf(w, string(data))
}

func proveConnect(w http.ResponseWriter, r *http.Request){
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data *ProveConnection
	decoder:= json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(data)
	user,err:=db_work.GetUser("login" , map[string]string{"login":data.Login, "pass":data.Pass})
	if err!=nil{
		fmt.Fprintf(w, "Error")
		return
	}
	fmt.Println(user.ID)
	//fmt.Println(&data.Login)
	//fmt.Println(&data.Pass)
	fmt.Fprintf(w, "Connect")
}

func getMessages(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data RequestGetMessage
	decoder:= json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(data.Chat_Id, data.Author)
	//fmt.Print(params.Get("Author"))
	id := data.Chat_Id
	for _,r := range messagesBlock{
		if r.Chat_Id == id{
			need_chat_messages := *r
			data,_ := json.Marshal(need_chat_messages.Messages)
			//fmt.Fprintf(w, string(data))
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		}
	}
	errAnswer := ErrorAnswer{"Error", "Chat is undefined"}
	js,err := json.Marshal(errAnswer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return
}

func testDb(w http.ResponseWriter, r *http.Request){
	now:=db_work.GetInfo()
	model :=models.GetModels()
	//defer r.Body.Close()
	fmt.Print(now, model)
	w.Write([]byte(now+model))
	return
}

func ApiRouter(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars:=mux.Vars(r)
	api.MainApiRouter(vars["key"], vars["var1"], w, r)
}


func main(){
	//for i := 1; i < 3; i++ {
	//	createMainChat(float64(i))
	//}
	go broadcaster()
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Handle("/ws", websocket.Handler(SocketListener))
	myRouter.HandleFunc("/proveConnect", proveConnect)
	myRouter.HandleFunc("/testDb", testDb)
	myRouter.HandleFunc("/getChats", getChats)
	myRouter.HandleFunc("/getMessages", getMessages)
	myRouter.HandleFunc("/api/{key}/{var1}", ApiRouter)
	//if err := myRouter.ListenAndServe(":1234", nil); err != nil {
	//	log.Fatal("ListenAndServe:", err)
	//}
	log.Fatal(http.ListenAndServe(":1234", myRouter))
}





