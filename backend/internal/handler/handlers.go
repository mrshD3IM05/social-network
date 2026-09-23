package handler

import (
	"sn-backend/internal/handler/authhandler"
	"sn-backend/internal/handler/commenthandler"
	"sn-backend/internal/handler/filehandler"
	"sn-backend/internal/handler/grouphandler"
	"sn-backend/internal/handler/posthandler"
	"sn-backend/internal/handler/userhandler"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/authsvc"
	"sn-backend/internal/service/commentsvc"
	"sn-backend/internal/service/eventsvc"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/followsvc"
	"sn-backend/internal/service/groupsvc"
	"sn-backend/internal/service/postsvc"
	"sn-backend/internal/service/sessionsvc"
	"sn-backend/internal/service/usersvc"
	ws "sn-backend/internal/websocket"
)

type Handlers struct {
	Auth      *authhandler.Handler
	User      *userhandler.Handler
	Post      *posthandler.Handler
	Comment   *commenthandler.Handler
	File      *filehandler.Handler
	Group     *grouphandler.Handler
	WebSocket *ws.Hub
}

func New(repo *repository.Repository) *Handlers {
	webSocket := ws.NewHub(repo)
	postService := postsvc.New(repo)
	return &Handlers{
		Auth:      authhandler.New(authsvc.New(repo), sessionsvc.New(repo), webSocket),
		User:      userhandler.New(usersvc.New(repo), sessionsvc.New(repo), followsvc.New(repo)),
		Post:      posthandler.New(postService, sessionsvc.New(repo)),
		Comment:   commenthandler.New(commentsvc.New(repo, webSocket), sessionsvc.New(repo)),
		File:      filehandler.New(filesvc.New(repo, "uploads"), sessionsvc.New(repo)),
		Group:     grouphandler.New(groupsvc.New(repo, webSocket), postService, eventsvc.New(repo, webSocket), sessionsvc.New(repo)),
		WebSocket: webSocket,
	}
}
