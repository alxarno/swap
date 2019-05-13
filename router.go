package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/robbert229/jwt"
	db "github.com/swap-messenger/swap/db2"
	"github.com/swap-messenger/swap/settings"
	api "github.com/swap-messenger/swap/src/api"
	engine "github.com/swap-messenger/swap/src/messages"
	"golang.org/x/net/websocket"
	// "google.golang.org/genproto/protobuf/api"
)

func apiRouter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(r)
	api.Api(vars["key"], vars["var1"], w, r)
}

func stand(w http.ResponseWriter, r *http.Request) {
	file := "./frontend/index.html"
	http.ServeFile(w, r, file)
	return
}

func static(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["key3"] != "main.css" {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "accept-encoding")
		file := "./frontend/" + vars["key1"] + "/" + vars["key2"] + "/" + vars["key3"] + ".gz"
		http.ServeFile(w, r, file)
		return
	} else {
		file := "./frontend/" + vars["key1"] + "/" + vars["key2"] + "/" + vars["key3"]
		http.ServeFile(w, r, file)
		return
	}

}

func staticNotGzip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := "./frontend/static/" + vars["key2"] + "/" + vars["key3"]
	http.ServeFile(w, r, file)
	return

}

func logos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := "./frontend/" + vars["key1"]
	http.ServeFile(w, r, file)
	return
}

func fonts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := "./frontend/" + vars["key1"] + "/" + vars["key2"]
	http.ServeFile(w, r, file)
	return
}

func proveConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var data struct {
		Login string `json:"login"`
		Pass  string `json:"pass"`
	}
	// var data *ProveConnection
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil {
		log.Println(err)
	}
	_, err = db.GetUserByLoginAndPass(data.Login, data.Pass)
	if err != nil {
		fmt.Fprintf(w, "Error")
		return
	}

	fmt.Fprintf(w, "Connect")
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	sett, err := settings.GetSettings()
	if err != nil {
		w.Write([]byte("Error with security"))
		return
	}
	secret := sett.Backend.SecretKeyForToken
	vars := mux.Vars(r)
	algorithm := jwt.HmacSha256(secret)
	// Gologer.PInfo(vars["link"])

	claims, err := algorithm.Decode(vars["link"])
	if err != nil {
		w.Write([]byte("Fail decode link"))
	}
	nTime, err := claims.Get("time")
	if err != nil {
		w.Write([]byte("Fail get time"))
	}
	path, err := claims.Get("path")
	if err != nil {
		w.Write([]byte("Fail get path"))
	}
	sPath := path.(string)
	iTime := nTime.(float64)
	if int64(iTime) < time.Now().Unix() {
		w.Write([]byte("Link is unavailable"))
	}
	// Gologer.PInfo(sPath)
	file := "./public/files/" + sPath
	http.ServeFile(w, r, file)
}

func newRouter() *mux.Router {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", stand)
	myRouter.HandleFunc("/login", stand)
	myRouter.HandleFunc("/reg", stand)
	myRouter.HandleFunc("/messages", stand)
	myRouter.HandleFunc("/messages/{key}", stand)
	myRouter.HandleFunc("/getFile/{link}/{name}", downloadFile)
	myRouter.Handle("/ws", websocket.Handler(engine.ConnectionHandler))
	myRouter.HandleFunc("/proveConnect", proveConnect)
	myRouter.HandleFunc("/api/{key}/{var1}", apiRouter)
	myRouter.HandleFunc("/{key1}", logos)
	myRouter.HandleFunc("/{key1}/{key2}", fonts)
	myRouter.HandleFunc("/staticingzip/{key2}/{key3}", staticNotGzip)
	myRouter.HandleFunc("/{key1}/{key2}/{key3}", static)

	return myRouter
}
