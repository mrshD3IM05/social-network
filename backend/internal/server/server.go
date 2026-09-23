package server

import (
	"net/http"

	handlers "sn-backend/internal/handler"
	"sn-backend/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, h *handlers.Handlers) {
	auth := middleware.NewAuth(h.Auth.Session)

	//Auth routes
	mux.Handle("POST /register", auth.Guest(http.HandlerFunc(h.Auth.Register)))
	mux.Handle("POST /login", auth.Guest(http.HandlerFunc(h.Auth.Login)))
	mux.Handle("POST /logout", auth.Authorized(http.HandlerFunc(h.Auth.Logout)))
	mux.Handle("GET /me", auth.Authorized(http.HandlerFunc(h.Auth.Me)))

	// user routes
	mux.Handle("GET /user/{id}", auth.Authorized(http.HandlerFunc(h.User.GetUser)))
	mux.Handle("POST /users/{id}/follow", auth.Authorized(http.HandlerFunc(h.User.FollowUser)))
	mux.Handle("DELETE /users/{id}/follow", auth.Authorized(http.HandlerFunc(h.User.UnfollowUser)))
	mux.Handle("POST /follow-requests/{id}/accept", auth.Authorized(http.HandlerFunc(h.User.RespondFollow)))
	mux.Handle("POST /follow-requests/{id}/decline", auth.Authorized(http.HandlerFunc(h.User.RespondFollow)))

	// post routes
	mux.Handle("GET /posts", auth.Authorized(http.HandlerFunc(h.Post.ListPosts)))
	mux.Handle("POST /posts", auth.Authorized(http.HandlerFunc(h.Post.CreatePost)))
	mux.Handle("PUT /posts/{id}", auth.Authorized(http.HandlerFunc(h.Post.UpdatePost)))
	mux.Handle("DELETE /posts/{id}", auth.Authorized(http.HandlerFunc(h.Post.DeletePost)))

	// comment routes (visibility follows the post: privacy rules or group membership)
	mux.Handle("GET /posts/{id}/comments", auth.Authorized(http.HandlerFunc(h.Comment.ListComments)))
	mux.Handle("POST /posts/{id}/comments", auth.Authorized(http.HandlerFunc(h.Comment.CreateComment)))
	mux.Handle("GET /posts/{id}", auth.Authorized(http.HandlerFunc(h.Post.GetPost)))

	// reaction routes
	mux.Handle("POST /posts/{id}/reactions", auth.Authorized(http.HandlerFunc(h.Post.ReactionPost)))
	mux.Handle("DELETE /posts/{id}/reactions", auth.Authorized(http.HandlerFunc(h.Post.DeleteReaction)))

	// file routes
	mux.Handle("POST /files", auth.Authorized(http.HandlerFunc(h.File.Upload)))
	mux.Handle("POST /avatar", auth.Authorized(http.HandlerFunc(h.File.SetAvatar)))
	mux.Handle("GET /fs/{id}", auth.Authorized(http.HandlerFunc(h.File.Download)))

	// group routes
	mux.Handle("POST /groups", auth.Authorized(http.HandlerFunc(h.Group.CreateGroup)))
	mux.Handle("GET /groups", auth.Authorized(http.HandlerFunc(h.Group.ListGroups)))
	mux.Handle("GET /groups/{id}", auth.Authorized(http.HandlerFunc(h.Group.GetGroup)))
	mux.Handle("GET /groups/{id}/members", auth.Authorized(http.HandlerFunc(h.Group.GetGroupMembers)))
	mux.Handle("POST /groups/{id}/invitations", auth.Authorized(http.HandlerFunc(h.Group.InviteUser)))
	mux.Handle("POST /group-invitations/{id}/accept", auth.Authorized(http.HandlerFunc(h.Group.RespondInvitation)))
	mux.Handle("POST /group-invitations/{id}/decline", auth.Authorized(http.HandlerFunc(h.Group.RespondInvitation)))
	mux.Handle("GET /group-invitations", auth.Authorized(http.HandlerFunc(h.Group.PendingInvitations)))
	mux.Handle("POST /groups/{id}/join-requests", auth.Authorized(http.HandlerFunc(h.Group.RequestJoin)))
	mux.Handle("POST /group-join-requests/{id}/accept", auth.Authorized(http.HandlerFunc(h.Group.RespondJoinRequest)))
	mux.Handle("POST /group-join-requests/{id}/decline", auth.Authorized(http.HandlerFunc(h.Group.RespondJoinRequest)))
	mux.Handle("GET /groups/{id}/join-requests", auth.Authorized(http.HandlerFunc(h.Group.PendingJoinRequests)))

	// group post routes (members only, enforced in the services)
	mux.Handle("GET /groups/{id}/posts", auth.Authorized(http.HandlerFunc(h.Group.ListGroupPosts)))
	mux.Handle("POST /groups/{id}/posts", auth.Authorized(http.HandlerFunc(h.Group.CreateGroupPost)))

	// websocket routes
	mux.Handle("GET /ws", auth.Authorized(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.WebSocket.ServeHTTP(w, r, h.Auth.Session)
	})))
}
