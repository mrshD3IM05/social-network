package server

import (
	"net/http"

	handlers "sn-backend/internal/handler"
	"sn-backend/internal/handler/messagehandler"
	"sn-backend/internal/middleware"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/messagesvc"
)

// RegisterMessageRoutes adds the direct message endpoints.
// It lives in its own file so the main route table stays untouched.
func RegisterMessageRoutes(mux *http.ServeMux, repo *repository.Repository, h *handlers.Handlers) {
	auth := middleware.NewAuth(h.Auth.Session)
	messages := messagehandler.New(
		messagesvc.New(repo),
		filesvc.New(repo, "uploads"),
		h.Auth.Session,
		h.WebSocket,
	)

	mux.Handle("GET /messages/{id}", auth.Authorized(http.HandlerFunc(messages.History)))
	mux.Handle("POST /messages", auth.Authorized(http.HandlerFunc(messages.Send)))
}
