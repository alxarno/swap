package main

import (
	"encoding/json"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/alxarno/swap/src/api"

	logger "github.com/alxarno/swap/logger"
	"github.com/alxarno/swap/settings"
	engine "github.com/alxarno/swap/src/messages"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
	"github.com/robbert229/jwt"
	"golang.org/x/net/websocket"
)

type route func(string, func(http.ResponseWriter, *http.Request), ...string)
type subroute func(string) *api.Router
type middlewareFunc func(http.Handler) http.Handler

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

func logginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func AdditionalHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(path.Base(r.URL.Path)) == ".js" {
			w.Header().Add("Content-Type", "application/javascript")
		}
		next.ServeHTTP(w, r)
	})
}

func _CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Auth-Token")
		next.ServeHTTP(w, r)
	})
}

func newRouter() *mux.Router {
	box := packr.NewBox("./ui")
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Handle("/", http.FileServer(box))
	myRouter.HandleFunc("/info", info)
	myRouter.HandleFunc("/getFile/{link}/{name}", downloadFile)
	myRouter.Handle("/ws", websocket.Handler(engine.ConnectionHandler))
	api.RegisterEndpoints(newSubRoute(myRouter)("/api"))

	myRouter.Handle("/{key1}", http.FileServer(box))
	myRouter.Handle("/{key1}/{key2}", http.FileServer(box))

	myRouter.Use(logginMiddleware)
	if settings.ServiceSettings.Service.CORS {
		myRouter.Use(_CORSMiddleware)
	}
	myRouter.Use(AdditionalHeaders)

	return myRouter
}

func newRoute(router *mux.Router) route {
	return func(pattern string, handler func(w http.ResponseWriter, r *http.Request), methods ...string) {
		r := (*router).HandleFunc(pattern, handler)
		if len(methods) > 1 {
			r.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
				// For Cross Domain requests need additional OPTIONS request
				if settings.ServiceSettings.Service.CORS && r.Method == http.MethodOptions {
					return true
				}
				for _, v := range methods {
					if v == r.Method {
						return true
					}
				}
				return false
			})
		} else {
			if settings.ServiceSettings.Service.CORS {
				r.Methods(methods[0], "OPTIONS")
			} else {
				r.Methods(methods[0])
			}
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
