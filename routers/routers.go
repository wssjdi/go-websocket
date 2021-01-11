package routers

import (
	"github.com/woodylan/go-websocket/api/bind2group"
	"github.com/woodylan/go-websocket/api/closeclient"
	"github.com/woodylan/go-websocket/api/getonlinelist"
	"github.com/woodylan/go-websocket/api/getuserclients"
	"github.com/woodylan/go-websocket/api/register"
	"github.com/woodylan/go-websocket/api/send2client"
	"github.com/woodylan/go-websocket/api/send2clients"
	"github.com/woodylan/go-websocket/api/send2group"
	"github.com/woodylan/go-websocket/api/send2user"
	"github.com/woodylan/go-websocket/servers"
	"io"
	"net/http"
)

func Init() {

	http.HandleFunc("/health", Health)

	//Rest Api
	registerHandler := &register.Controller{}
	sendToClientHandler := &send2client.Controller{}
	sendToClientsHandler := &send2clients.Controller{}
	sendToGroupHandler := &send2group.Controller{}
	bindToGroupHandler := &bind2group.Controller{}
	sendToUserHandler := &send2user.Controller{}
	getGroupListHandler := &getonlinelist.Controller{}
	getUserClientsHandler := &getuserclients.Controller{}
	closeClientHandler := &closeclient.Controller{}

	http.HandleFunc("/api/register", registerHandler.Run)
	http.HandleFunc("/api/bind/2/group", AccessTokenMiddleware(bindToGroupHandler.Run))
	http.HandleFunc("/api/group/list", AccessTokenMiddleware(getGroupListHandler.Run))
	http.HandleFunc("/api/user/list", AccessTokenMiddleware(getUserClientsHandler.Run))
	http.HandleFunc("/api/send/2/client", AccessTokenMiddleware(sendToClientHandler.Run))
	http.HandleFunc("/api/send/2/clients", AccessTokenMiddleware(sendToClientsHandler.Run))
	http.HandleFunc("/api/send/2/group", AccessTokenMiddleware(sendToGroupHandler.Run))
	http.HandleFunc("/api/send/2/user", AccessTokenMiddleware(sendToUserHandler.Run))
	http.HandleFunc("/api/close/client", AccessTokenMiddleware(closeClientHandler.Run))

	//WebSocket Api
	websocketHandler := &servers.Controller{}
	http.HandleFunc("/ws", websocketHandler.Run)

	servers.StartWebSocket()

	go servers.WriteMessage()
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = io.WriteString(w, "OK")
	return
}
