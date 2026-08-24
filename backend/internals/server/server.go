package server

import (
	"net/http"

	handlers "sn-backend/internals/handler"
	"sn-backend/internals/middleware"
)

func RegisterRoutes(mux *http.ServeMux, h *handlers.Handlers) {
	auth := middleware.NewAuth(h.Auth.Session)

	//Auth routes
	mux.Handle("POST /register", auth.Guest(http.HandlerFunc(h.Auth.Register)))
	mux.Handle("POST /login", auth.Guest(http.HandlerFunc(h.Auth.Login)))
	mux.Handle("POST /logout", auth.Authorized(http.HandlerFunc(h.Auth.Logout)))
	mux.Handle("GET /me", auth.Authorized(http.HandlerFunc(h.Auth.Me)))

	// user routes
	mux.Handle("GET /users", auth.Authorized(http.HandlerFunc(h.User.ListUsers)))
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

	// file routes
	mux.Handle("POST /files", auth.Authorized(http.HandlerFunc(h.File.Upload)))
	mux.Handle("GET /fs/{id}", auth.Authorized(http.HandlerFunc(h.File.Download)))

	// websocket routes
	mux.Handle("GET /ws", auth.Authorized(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.WebSocket.ServeHTTP(w, r, h.Auth.Session)
	})))
}
