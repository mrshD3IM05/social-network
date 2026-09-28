package server

import (
	"net/http"

	handlers "sn-backend/internal/handler"
	"sn-backend/internal/middleware"
)

func RegisterRoutes(mux *http.ServeMux, h *handlers.Handlers) {
	auth := middleware.NewAuth(h.Auth.Session)

	//Auth routes
	mux.Handle("POST /register", middleware.AuthRateLimit(auth.Guest(http.HandlerFunc(h.Auth.Register))))
	mux.Handle("POST /login", middleware.AuthRateLimit(auth.Guest(http.HandlerFunc(h.Auth.Login))))
	mux.Handle("POST /logout", auth.Authorized(http.HandlerFunc(h.Auth.Logout)))
	mux.Handle("GET /me", auth.Authorized(http.HandlerFunc(h.Auth.Me)))

	// user routes
	mux.Handle("GET /users", auth.Authorized(http.HandlerFunc(h.User.ListUsers)))
	mux.Handle("GET /user/{id}", auth.Authorized(http.HandlerFunc(h.User.GetUser)))
	mux.Handle("GET /users/{id}/followers", auth.Authorized(http.HandlerFunc(h.User.Followers)))
	mux.Handle("GET /users/{id}/following", auth.Authorized(http.HandlerFunc(h.User.Following)))
	mux.Handle("GET /users/{id}/follow", auth.Authorized(http.HandlerFunc(h.User.FollowStatus)))
	mux.Handle("POST /users/{id}/follow", auth.Authorized(http.HandlerFunc(h.User.FollowUser)))
	mux.Handle("DELETE /users/{id}/follow", auth.Authorized(http.HandlerFunc(h.User.UnfollowUser)))
	mux.Handle("POST /follow-requests/{id}/accept", auth.Authorized(http.HandlerFunc(h.User.RespondFollow)))
	mux.Handle("POST /follow-requests/{id}/decline", auth.Authorized(http.HandlerFunc(h.User.RespondFollow)))
	mux.Handle("PUT /me/privacy", auth.Authorized(http.HandlerFunc(h.User.SetPrivacy)))

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
	mux.Handle("PUT /groups/{id}", auth.Authorized(http.HandlerFunc(h.Group.UpdateGroup)))
	mux.Handle("DELETE /groups/{id}", auth.Authorized(http.HandlerFunc(h.Group.DeleteGroup)))
	mux.Handle("POST /groups/{id}/avatar", auth.Authorized(http.HandlerFunc(h.Group.SetGroupAvatar)))
	mux.Handle("GET /groups/{id}/members", auth.Authorized(http.HandlerFunc(h.Group.GetGroupMembers)))
	mux.Handle("DELETE /groups/{id}/members/{userID}", auth.Authorized(http.HandlerFunc(h.Group.RemoveMember)))
	mux.Handle("GET /groups/{id}/messages", auth.Authorized(http.HandlerFunc(h.Group.ListMessages)))
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
	mux.Handle("DELETE /groups/{id}/posts/{post_id}", auth.Authorized(http.HandlerFunc(h.Group.DeleteGroupPost)))

	// group event routes (members only, enforced in the services)
	mux.Handle("GET /groups/{id}/events", auth.Authorized(http.HandlerFunc(h.Group.ListEvents)))
	mux.Handle("POST /groups/{id}/events", auth.Authorized(http.HandlerFunc(h.Group.CreateEvent)))
	mux.Handle("POST /events/{id}/response", auth.Authorized(http.HandlerFunc(h.Group.RespondEvent)))
	mux.Handle("GET /events/{id}/response", auth.Authorized(http.HandlerFunc(h.Group.MyEventResponse)))

	// direct message routes (the sender must follow, or be followed by, the recipient)
	mux.Handle("GET /messages/{id}", auth.Authorized(http.HandlerFunc(h.Message.History)))
	mux.Handle("POST /messages", auth.Authorized(http.HandlerFunc(h.Message.Send)))

	// websocket routes
	mux.Handle("GET /ws", auth.Authorized(h.WebSocket))
}
