package server

import (
	"log"
	"net/http"
	"time"

	handlers "sn-backend/internal/handler"
	"sn-backend/internal/handler/commenthandler"
	"sn-backend/internal/handler/common"
	"sn-backend/internal/handler/messagehandler"
	"sn-backend/internal/handler/notificationhandler"
	"sn-backend/internal/handler/posthandler"
	"sn-backend/internal/handler/socialhandler"
	"sn-backend/internal/middleware"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/commentsvc"
	"sn-backend/internal/service/messagesvc"
	"sn-backend/internal/service/notificationsvc"
	"sn-backend/internal/service/postsvc"
	"sn-backend/internal/service/socialsvc"
)

// followNotifier lets the follow service persist a notification through the
// repository and push it through the websocket hub in one dependency.
type followNotifier struct {
	repo *repository.Repository
	hub  interface{ PublishNotification(*model.Notification) }
}

func (n followNotifier) CreateNotification(notification *model.Notification) error {
	return n.repo.CreateNotification(notification)
}

func (n followNotifier) PublishNotification(notification *model.Notification) {
	n.hub.PublishNotification(notification)
}

// RegisterRoutesExtra adds the endpoints the original route table was missing:
// comments, private-post audiences, follower lists, pending follow requests,
// the profile privacy toggle, notification history and message history.
// It is kept in its own file so the original RegisterRoutes stays untouched.
func RegisterRoutesExtra(mux *http.ServeMux, repo *repository.Repository, h *handlers.Handlers) {
	auth := middleware.NewAuth(h.Auth.Session)
	sessions := h.Auth.Session

	// Follow requests now raise a notification for the target user.
	h.User.Follow.WithNotifier(followNotifier{repo: repo, hub: h.WebSocket})

	social := socialhandler.New(socialsvc.New(repo), sessions)
	comments := commenthandler.New(commentsvc.New(repo), h.File.Service, sessions)
	notifications := notificationhandler.New(notificationsvc.New(repo), sessions)
	messages := messagehandler.New(messagesvc.New(repo), sessions)
	audience := posthandler.NewAudience(postsvc.NewAudience(repo), sessions)

	// health, for the container healthcheck (no session needed)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		common.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// people and profiles
	mux.Handle("GET /users", auth.Authorized(http.HandlerFunc(social.Users)))
	mux.Handle("GET /users/{id}/followers", auth.Authorized(http.HandlerFunc(social.Followers)))
	mux.Handle("GET /users/{id}/following", auth.Authorized(http.HandlerFunc(social.Following)))
	mux.Handle("GET /users/{id}/posts", auth.Authorized(http.HandlerFunc(social.UserPosts)))
	mux.Handle("GET /users/{id}/relationship", auth.Authorized(http.HandlerFunc(social.Relationship)))
	mux.Handle("PATCH /me", auth.Authorized(http.HandlerFunc(social.UpdateMe)))

	// follow requests
	mux.Handle("GET /follow-requests", auth.Authorized(http.HandlerFunc(social.PendingFollowRequests)))

	// comments
	mux.Handle("GET /posts/{id}/comments", auth.Authorized(http.HandlerFunc(comments.List)))
	mux.Handle("POST /posts/{id}/comments", auth.Authorized(http.HandlerFunc(comments.Create)))
	mux.Handle("DELETE /comments/{id}", auth.Authorized(http.HandlerFunc(comments.Delete)))
	mux.Handle("POST /comments/{id}/reactions", auth.Authorized(http.HandlerFunc(comments.React)))
	mux.Handle("DELETE /comments/{id}/reactions", auth.Authorized(http.HandlerFunc(comments.Unreact)))

	// chosen followers of a private post
	mux.Handle("GET /posts/{id}/audience", auth.Authorized(http.HandlerFunc(audience.Get)))
	mux.Handle("PUT /posts/{id}/audience", auth.Authorized(http.HandlerFunc(audience.Set)))

	// notifications
	mux.Handle("GET /notifications", auth.Authorized(http.HandlerFunc(notifications.List)))
	mux.Handle("GET /notifications/unread", auth.Authorized(http.HandlerFunc(notifications.UnreadCount)))
	mux.Handle("POST /notifications/read", auth.Authorized(http.HandlerFunc(notifications.MarkAllRead)))
	mux.Handle("POST /notifications/{id}/read", auth.Authorized(http.HandlerFunc(notifications.MarkRead)))

	// chat history
	mux.Handle("GET /conversations", auth.Authorized(http.HandlerFunc(messages.Inbox)))
	mux.Handle("GET /messages/{id}", auth.Authorized(http.HandlerFunc(messages.Conversation)))
}

// StartSessionCleanup reaps expired session rows in the background; previously
// they were only deleted when someone presented the stale cookie.
func StartSessionCleanup(repo *repository.Repository, every time.Duration) {
	go func() {
		for {
			if removed, err := repo.DeleteExpiredSessions(); err != nil {
				log.Printf("session cleanup: %v", err)
			} else if removed > 0 {
				log.Printf("session cleanup: removed %d expired sessions", removed)
			}
			time.Sleep(every)
		}
	}()
}
