package handler

import (
	"sn-backend/internal/handler/authhandler"
	"sn-backend/internal/handler/commenthandler"
	"sn-backend/internal/handler/filehandler"
	"sn-backend/internal/handler/grouphandler"
	"sn-backend/internal/handler/messagehandler"
	"sn-backend/internal/handler/posthandler"
	"sn-backend/internal/handler/userhandler"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/authsvc"
	"sn-backend/internal/service/commentsvc"
	"sn-backend/internal/service/eventsvc"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/followsvc"
	"sn-backend/internal/service/groupsvc"
	"sn-backend/internal/service/messagesvc"
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
	Message   *messagehandler.Handler
	WebSocket *ws.Hub
}

// New builds every handler once, so RegisterRoutes only has to wire paths to
// methods. Each service is created a single time and shared.
func New(repo *repository.Repository) *Handlers {
	session := sessionsvc.New(repo)
	postService := postsvc.New(repo)
	fileService := filesvc.New(repo, "uploads")
	webSocket := ws.NewHub(repo, session)
	return &Handlers{
		Auth:      authhandler.New(authsvc.New(repo), session, webSocket),
		User:      userhandler.New(usersvc.New(repo), session, followsvc.New(repo, webSocket), postService),
		Post:      posthandler.New(postService, session),
		Comment:   commenthandler.New(commentsvc.New(repo, webSocket), session),
		File:      filehandler.New(fileService, session),
		Group:     grouphandler.New(groupsvc.New(repo, webSocket), postService, eventsvc.New(repo, webSocket), fileService, session),
		Message:   messagehandler.New(messagesvc.New(repo), fileService, session, webSocket),
		WebSocket: webSocket,
	}
}
