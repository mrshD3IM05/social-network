package handler

import (
	"sn-backend/internals/handler/authhandler"
	"sn-backend/internals/handler/filehandler"
	"sn-backend/internals/handler/posthandler"
	"sn-backend/internals/handler/userhandler"
	"sn-backend/internals/repository"
	"sn-backend/internals/service/authsvc"
	"sn-backend/internals/service/filesvc"
	"sn-backend/internals/service/followsvc"
	"sn-backend/internals/service/postsvc"
	"sn-backend/internals/service/sessionsvc"
	"sn-backend/internals/service/usersvc"
	ws "sn-backend/internals/websocket"
)

type Handlers struct {
	Auth      *authhandler.Handler
	User      *userhandler.Handler
	Post      *posthandler.Handler
	File      *filehandler.Handler
	WebSocket *ws.Hub
}

func New(repo *repository.Repository) *Handlers {
	webSocket := ws.NewHub(repo)
	return &Handlers{
		Auth:      authhandler.New(authsvc.New(repo), sessionsvc.New(repo), webSocket),
		User:      userhandler.New(usersvc.New(repo), sessionsvc.New(repo), followsvc.New(repo)),
		Post:      posthandler.New(postsvc.New(repo), sessionsvc.New(repo)),
		File:      filehandler.New(filesvc.New(repo, "uploads"), sessionsvc.New(repo)),
		WebSocket: webSocket,
	}
}
