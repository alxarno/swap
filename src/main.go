package main

import (
	"fmt"
	"net/http"
	"log"
	"golang.org/x/net/websocket"
	"encoding/json"
	"github.com/robbert229/jwt"
	"github.com/Spatium-Messenger/Server/models"
	api "github.com/Spatium-Messenger/Server/src/api2"
	dbApi "github.com/Spatium-Messenger/Server/db_api"
	"github.com/gorilla/mux"
	"time"
	engine "github.com/Spatium-Messenger/Server/src/message_engine"
	"net"
	"os"
	"path/filepath"
	"bufio"
	"github.com/Spatium-Messenger/Server/settings"
)
var (
	secret = settings.ServiceSettings.Server.SecretKeyForToken
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
	_,err = dbApi.GetUser("login" , map[string]interface{}{"login":data.Login, "pass":data.Pass})
	if err!=nil{
		fmt.Fprintf(w, "Error")
		return
	}

	fmt.Fprintf(w, "Connect")
}



func stand(w http.ResponseWriter, r *http.Request){
	file := "./frontend/index.html"
	http.ServeFile(w, r, file)
	return
}

func static(w http.ResponseWriter, r *http.Request){

	vars:=mux.Vars(r)

	if vars["key3"] != "main.css" {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "accept-encoding")
		file := "./frontend/" + vars["key1"] + "/" + vars["key2"] + "/" + vars["key3"] + ".gz"
		http.ServeFile(w, r,file)
		return
	}else{
		file := "./frontend/" + vars["key1"] + "/" + vars["key2"] + "/" + vars["key3"]
		http.ServeFile(w, r,file)
		return
	}


}

func staticNotGzip(w http.ResponseWriter, r *http.Request){
	//w.Header().Set("Content-Encoding", "gzip")
	//w.Header().Set("Vary", "accept-encoding")
	vars:=mux.Vars(r)
	//file := "./frontend/"+vars["key1"]+"/"+vars["key2"]+"/"+vars["key3"]+".gz"
	//if vars["key3"] == "main.css"{
		file := "./frontend/static/"+vars["key2"]+"/"+vars["key3"]
		http.ServeFile(w, r,file)
		return

}

func logos(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	file := "./frontend/"+vars["key1"]
	http.ServeFile(w, r, file)
	return
}

func fonts(w http.ResponseWriter, r *http.Request){
	vars:=mux.Vars(r)
	file := "./frontend/"+vars["key1"]+"/"+vars["key2"]
	//w.Header().Set("Content-Encoding", "gzip")
	//w.Header().Set("Vary", "accept-encoding")
	http.ServeFile(w, r, file)
	return
}

func ApiRouter(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars:=mux.Vars(r)
	api.Api(vars["key"], vars["var1"], w, r)
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
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

func redirectToHttps(w http.ResponseWriter, r *http.Request) {
	// Redirect the incoming HTTP request. Note that "127.0.0.1:8081" will only work if you are accessing the server from your local machine.
	http.Redirect(w, r, "https://192.168.56.1:1235"+r.RequestURI, http.StatusMovedPermanently)
}

func printLogo(){
	fmt.Println()
	fmt.Println(" ________  ________  ________  _________  ___  ___  ___  _____ ______      ")
	fmt.Println(`|\   ____\|\   __  \|\   __  \|\___   ___|\  \|\  \|\  \|\   _ \  _   \    `)
	fmt.Println(`\ \  \___|\ \  \|\  \ \  \|\  \|___ \  \_\ \  \ \  \\\  \ \  \\\__\ \  \   `)
	fmt.Println(` \ \_____  \ \   ____\ \   __  \   \ \  \ \ \  \ \  \\\  \ \  \\|__| \  \  `)
	fmt.Println(`  \|____|\  \ \  \___|\ \  \ \  \   \ \  \ \ \  \ \  \\\  \ \  \    \ \  \ `)
	fmt.Println(`    ____\_\  \ \__\    \ \__\ \__\   \ \__\ \ \__\ \_______\ \__\    \ \__\`)
	fmt.Println(`   |\_________\|__|     \|__|\|__|    \|__|  \|__|\|_______|\|__|     \|__|`)
	fmt.Println(`   \|_________|                                                            `)
	fmt.Println()
}



func main(){
	dbApi.BeginDB()
	//go broadcaster()
	engine.StartCoreMessenger()

	RemoveContents("./public/files")
	os.MkdirAll("./public/files/min", os.ModePerm)
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", stand)
	myRouter.HandleFunc("/login", stand)
	myRouter.HandleFunc("/reg", stand)
	myRouter.HandleFunc("/messages", stand)
	myRouter.HandleFunc("/messages/{key}", stand)
	myRouter.HandleFunc("/getFile/{link}/{name}", downloadFile)
	myRouter.Handle("/ws", websocket.Handler(engine.SocketListener))
	myRouter.HandleFunc("/proveConnect", proveConnect)
	myRouter.HandleFunc("/api/{key}/{var1}", ApiRouter)
	myRouter.HandleFunc("/{key1}", logos)
	myRouter.HandleFunc("/{key1}/{key2}", fonts)
	myRouter.HandleFunc("/staticingzip/{key2}/{key3}",staticNotGzip)
	myRouter.HandleFunc("/{key1}/{key2}/{key3}",static)

	my_addres:= ""
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}
	err = settings.LoadSettings()

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				my_addres+=ipnet.IP.String()
				my_addres+=":"+settings.ServiceSettings.Server.Host+"\t"
			}
		}
	}


	if err!=nil{
		fmt.Println(err.Error())
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}
	printLogo()
	//err = dbApi

	//if err!=nil{
	//	fmt.Println(err.Error())
	//	bufio.NewReader(os.Stdin).ReadBytes('\n')
	//	return
	//}
	os.Stderr.WriteString("Spatium started on \t"+ my_addres+"\n")

	if settings.ServiceSettings.Server.Encryption{
		log.Fatal("ListenAndServeTLS: ",http.ListenAndServeTLS(
			":"+ settings.ServiceSettings.Server.Host,
			settings.ServiceSettings.Server.Cert_file,
			settings.ServiceSettings.Server.Key_file,
			myRouter))
	}else{
		log.Fatal("ListenAndServe: ", http.ListenAndServe(
		 	":"+ settings.ServiceSettings.Server.Host,
			 myRouter))
	}

	//http.ListenAndServe(":1234", http.HandlerFunc(redirectToHttps))


}





