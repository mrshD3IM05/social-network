package handler

import (
	"sn-backend/internal/handler/authhandler"
	"sn-backend/internal/handler/filehandler"
	"sn-backend/internal/handler/posthandler"
	"sn-backend/internal/handler/userhandler"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/authsvc"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/followsvc"
	"sn-backend/internal/service/postsvc"
	"sn-backend/internal/service/sessionsvc"
	"sn-backend/internal/service/usersvc"
	ws "sn-backend/internal/websocket"
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
