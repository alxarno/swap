package main

import (
	"fmt"
	"net/http"
	"log"
	"golang.org/x/net/websocket"
	"encoding/json"
	//"strconv"
	"github.com/robbert229/jwt"
	db_work "github.com/AlexArno/spatium/db_work"
	models "github.com/AlexArno/spatium/models"
	api "github.com/AlexArno/spatium/src/api"
	messages_work "github.com/AlexArno/spatium/src/messages"
	"github.com/gorilla/mux"
	"time"
	engine "github.com/AlexArno/spatium/src/message_engine"
	)
var (
	secret = "321312421"
	//Nmessages =engine.Messages
)

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


type client chan<-models.NewMessageToUser
var (
	chats []*models.Chat
	messagesBlock []*models.MessageBlock
	messages = make(chan models.NewMessageToUser)
	entering = make(chan client)
	leaving = make(chan client)
)


func broadcaster(){
	clients:= make(map[client]bool)
	for{
		select {
			case msg:=<-messages:
				for cli := range clients{
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

func writerUser(ws *websocket.Conn, ch<-chan  models.NewMessageToUser){
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
	ch:= make(chan  models.NewMessageToUser)
	go writerUser(ws, ch)
	entering<-ch
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive", err)
			break
		}

		send, err := messages_work.NewMessage(&reply)
		if err != nil{
			fmt.Println("Decode message", err)
			break
		}

		for v,r := range chats{
			if float64(r.ID) == *send.Chat_Id{
				chats[v].LastSender = *send.Author_Name
				chats[v].LastMessage = *send.Content.Message
			}
		}

		messages<-send
	}
	leaving<-ch
	ws.Close()
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
	//fmt.Println(data)
	_,err =db_work.GetUser("login" , map[string]string{"login":data.Login, "pass":data.Pass})
	if err!=nil{
		fmt.Fprintf(w, "Error")
		return
	}
	//fmt.Println(user.ID)
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
	//fmt.Println(data.Chat_Id, data.Author)
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
	//fmt.Print(now, model)
	w.Write([]byte(now+model))
	return
}

func ApiRouter(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars:=mux.Vars(r)
	api.MainApiRouter(vars["key"], vars["var1"], w, r)
}

func downloadFile(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	algorithm :=  jwt.HmacSha256(secret)
	claims, err := algorithm.Decode(vars["link"])
	if err != nil {
		w.Write([]byte("Fail decode link"))
	}
	n_time,err :=claims.Get("time")
	if err != nil{
		w.Write([]byte("Fail get time"))
	}
	path, err:= claims.Get("path")
	if err != nil{
		w.Write([]byte("Fail get path"))
	}
	s_path := path.(string)
	i_time := n_time.(float64)
	if  int64(i_time)<time.Now().Unix(){
		w.Write([]byte("Link is unavailable"))
	}
	file := "./public/files/"+s_path
	http.ServeFile(w,r,file)
}

func main(){
	db_work.OpenDB()
	//go broadcaster()
	engine.StartCoreMessenger()

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Handle("/ws", websocket.Handler(engine.SocketListener))
	myRouter.HandleFunc("/proveConnect", proveConnect)
	myRouter.HandleFunc("/testDb", testDb)
	//myRouter.HandleFunc("/getChats", getChats)
	myRouter.HandleFunc("/getMessages", getMessages)
	myRouter.HandleFunc("/api/{key}/{var1}", ApiRouter)
	myRouter.HandleFunc("/getFile/{link}/{name}", downloadFile)
	//if err := myRouter.ListenAndServe(":1234", nil); err != nil {
	//	log.Fatal("ListenAndServe:", err)
	//}
	log.Fatal(http.ListenAndServe(":1234", myRouter))
}





