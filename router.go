package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alxarno/swap/src/api"

	logger "github.com/alxarno/swap/logger"
	"github.com/alxarno/swap/settings"
	engine "github.com/alxarno/swap/src/messages"
	"github.com/gorilla/mux"
	"github.com/robbert229/jwt"
	"golang.org/x/net/websocket"
)

type route func(pattern string, handler func(w http.ResponseWriter, r *http.Request), methods ...string)
type subroute func(pattern string) *api.Router

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
	file := settings.ServiceSettings.Backend.FilesPath + sPath
	logger.Logger.Printf("%s downloading - %s \n", r.RemoteAddr, sPath)
	http.ServeFile(w, r, file)
}

func middleware(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return cors(logs(handler))
}

func logs(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if (*r).Method != "OPTIONS" {
			logger.Logger.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		}
		handler(w, r)
	}
}

func cors(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Auth-Token")
		handler(w, r)
	}
}

func newRouter() *mux.Router {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", middleware(stand))
	// myRouter.HandleFunc("/login", stand)
	// myRouter.HandleFunc("/reg", stand)
	myRouter.HandleFunc("/info", middleware(info))
	// myRouter.HandleFunc("/messages", stand)
	// myRouter.HandleFunc("/messages/{key}", stand)
	myRouter.HandleFunc("/getFile/{link}/{name}", middleware(downloadFile))
	myRouter.Handle("/ws", websocket.Handler(engine.ConnectionHandler))
	// myRouter.HandleFunc("/proveConnect", proveConnect)
	api.RegisterEndpoints(newSubRoute(myRouter)("/api"))
	// myRouter.HandleFunc("/{key1}", logos)
	// myRouter.HandleFunc("/{key1}/{key2}", fonts)
	// myRouter.HandleFunc("/staticingzip/{key2}/{key3}", staticNotGzip)
	// myRouter.HandleFunc("/{key1}/{key2}/{key3}", static)

	return myRouter
}

func newRoute(router *mux.Router) route {
	return func(pattern string, handler func(w http.ResponseWriter, r *http.Request), methods ...string) {
		r := (*router).HandleFunc(pattern, handler)
		if len(methods) > 1 {
			r.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
				for _, v := range methods {
					if v == r.Method {
						return true
					}
				}
				return false
			})
		} else {
			r.Methods(methods[0])
		}
	}
}

func newSubRoute(router *mux.Router) subroute {
	return func(pattern string) *api.Router {
		r := (*router).PathPrefix(pattern).Subrouter()
		return &api.Router{
			Route:    newRoute(r),
			Subroute: newSubRoute(r),
		}
	}
}
