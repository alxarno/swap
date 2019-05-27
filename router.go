package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alxarno/swap/settings"
	api "github.com/alxarno/swap/src/api"
	engine "github.com/alxarno/swap/src/messages"
	"github.com/gorilla/mux"
	"github.com/robbert229/jwt"
	"golang.org/x/net/websocket"
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

func info(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var answer = struct {
		Cert        bool  `json:"cert"`
		MaxFileSize int64 `json:"maxFileSize"`
	}{
		Cert:        true,
		MaxFileSize: settings.ServiceSettings.Service.MaxFileSize,
	}

	final, _ := json.Marshal(answer)
	w.Write(final)
}

// func proveConnect(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	var data proveConnection
// 	decoder := json.NewDecoder(r.Body)
// 	err := decoder.Decode(&data)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	_, err = db.GetUserByLoginAndPass(data.Login, data.Pass)
// 	if err != nil {
// 		fmt.Fprintf(w, "Error")
// 		return
// 	}

// 	fmt.Fprintf(w, "Connect")
// }

func downloadFile(w http.ResponseWriter, r *http.Request) {
	sett, err := settings.GetSettings()
	if err != nil {
		w.Write([]byte("Error with security"))
		return
	}
	secret := sett.Backend.SecretKeyForToken
	vars := mux.Vars(r)
	algorithm := jwt.HmacSha256(secret)

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
	file := settings.FilesDirPath + sPath
	http.ServeFile(w, r, file)
}

// type customWriter struct {
// 	w     http.ResponseWriter
// 	token string
// }

// func (s customWriter) Write(data []byte) (int, error) {
// 	key, err := engine.GetKeyByToken(s.token)
// 	if err != nil {
// 		return 0, errors.New("Cannot get user by key -> " + err.Error())
// 	}
// 	encMessage, err := swapcrypto.EncryptMessage(data, key)
// 	if err != nil {
// 		return 0, errors.New("Cannot encrypt data -> " + err.Error())
// 	}

// 	final, _ := json.Marshal(encMessage)
// 	return s.w.Write(final)
// }

// func (s customWriter) Header() http.Header {
// 	return s.w.Header()
// }

// func (s customWriter) WriteHeader(statusCode int) {
// 	s.w.WriteHeader(statusCode)
// }

// func customEncryption(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if !settings.ServiceSettings.Backend.Cert {
// 			token := r.Header.Get("Authorization")
// 			if token == "" {
// 				log.Println("Token is undefined -> ", token)
// 				return
// 			}
// 			var encryptedMessage models.EncryptedMessage
// 			if err := json.NewDecoder(r.Body).Decode(&encryptedMessage); err != nil {
// 				log.Println("Decode api call failed -> ", err.Error())
// 				return
// 			}
// 			message, err := swapcrypto.DecryptMessage(encryptedMessage.Key, encryptedMessage.IV, encryptedMessage.Data)
// 			if err != nil {
// 				log.Println("Decrypt api call failed -> ", err.Error())
// 				return
// 			}
// 			r.Body = ioutil.NopCloser(strings.NewReader(message))
// 			r.ContentLength = int64(len(message))
// 			cwriter := customWriter{w, token}
// 			handler(cwriter, r)
// 		} else {
// 			handler(w, r)
// 		}
// 	}
// }

func newRouter() *mux.Router {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", stand)
	// myRouter.HandleFunc("/login", stand)
	// myRouter.HandleFunc("/reg", stand)
	myRouter.HandleFunc("/info", info)
	// myRouter.HandleFunc("/messages", stand)
	// myRouter.HandleFunc("/messages/{key}", stand)
	myRouter.HandleFunc("/getFile/{link}/{name}", downloadFile)
	myRouter.Handle("/ws", websocket.Handler(engine.ConnectionHandler))
	// myRouter.HandleFunc("/proveConnect", proveConnect)
	myRouter.HandleFunc("/api/{key}/{var1}", apiRouter)
	// myRouter.HandleFunc("/{key1}", logos)
	// myRouter.HandleFunc("/{key1}/{key2}", fonts)
	// myRouter.HandleFunc("/staticingzip/{key2}/{key3}", staticNotGzip)
	// myRouter.HandleFunc("/{key1}/{key2}/{key3}", static)

	return myRouter
}
