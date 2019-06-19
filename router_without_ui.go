// +build !ui

package main

import (
	"github.com/alxarno/swap/settings"
	"github.com/alxarno/swap/src/api"
	engine "github.com/alxarno/swap/src/messages"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

func newRouter() *mux.Router {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/info", info)
	myRouter.HandleFunc("/getFile/{link}/{name}", downloadFile)
	myRouter.Handle("/ws", websocket.Handler(engine.ConnectionHandler))
	api.RegisterEndpoints(newSubRoute(myRouter)("/api"))

	myRouter.Use(logginMiddleware)
	if settings.ServiceSettings.Service.CORS {
		myRouter.Use(_CORSMiddleware)
	}
	myRouter.Use(AdditionalHeaders)

	return myRouter
}
