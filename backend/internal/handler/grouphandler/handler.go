package grouphandler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/eventsvc"
	"sn-backend/internal/service/groupsvc"
	"sn-backend/internal/service/postsvc"
	"sn-backend/internal/service/sessionsvc"
)

type Handler struct {
	Service *groupsvc.Service
	Post    *postsvc.Service
	Events  *eventsvc.Service
	Session *sessionsvc.Service
}

func New(service *groupsvc.Service, post *postsvc.Service, events *eventsvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Service: service, Post: post, Events: events, Session: session}
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	group, err := h.Service.Create(userID, r.FormValue("title"), r.FormValue("description"))
	if err != nil {
		switch {
		case errors.Is(err, groupsvc.ErrInvalidTitle), errors.Is(err, groupsvc.ErrInvalidDescription):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "could not create group", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusCreated, group)
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groups, err := h.Service.List(userID)
	if err != nil {
		http.Error(w, "could not list groups", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, groups)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	detail, err := h.Service.Detail(userID, groupID)
	if err != nil {
		writeError(w, err, "could not get group")
		return
	}
	common.WriteJSON(w, http.StatusOK, detail)
}

func (h *Handler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	members, err := h.Service.Members(userID, groupID)
	if err != nil {
		if errors.Is(err, groupsvc.ErrNotGroupMember) {
			http.Error(w, "only group members can view members", http.StatusForbidden)
			return
		}
		writeError(w, err, "could not get group members")
		return
	}
	common.WriteJSON(w, http.StatusOK, members)
}

func (h *Handler) InviteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	toUserID, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if err != nil || toUserID < 1 {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}
	invitation, err := h.Service.Invite(userID, groupID, toUserID)
	if err != nil {
		writeError(w, err, "could not invite user")
		return
	}
	common.WriteJSON(w, http.StatusCreated, invitation)
}

func (h *Handler) RespondInvitation(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	invitationID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid invitation id", http.StatusBadRequest)
		return
	}
	accept := r.URL.Path == "/group-invitations/"+r.PathValue("id")+"/accept"
	if err := h.Service.RespondInvitation(userID, invitationID, accept); err != nil {
		writeError(w, err, "could not respond to invitation")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RequestJoin(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	request, err := h.Service.RequestJoin(userID, groupID)
	if err != nil {
		writeError(w, err, "could not request to join group")
		return
	}
	common.WriteJSON(w, http.StatusCreated, request)
}

func (h *Handler) RespondJoinRequest(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	requestID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid join request id", http.StatusBadRequest)
		return
	}
	accept := r.URL.Path == "/group-join-requests/"+r.PathValue("id")+"/accept"
	if err := h.Service.RespondJoinRequest(userID, requestID, accept); err != nil {
		writeError(w, err, "could not respond to join request")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PendingInvitations(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	invitations, err := h.Service.PendingInvitations(userID)
	if err != nil {
		http.Error(w, "could not list pending invitations", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, invitations)
}

func (h *Handler) PendingJoinRequests(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	requests, err := h.Service.PendingJoinRequests(userID, groupID)
	if err != nil {
		if errors.Is(err, groupsvc.ErrNotGroupCreator) {
			http.Error(w, "only the group creator can view join requests", http.StatusForbidden)
			return
		}
		writeError(w, err, "could not list join requests")
		return
	}
	common.WriteJSON(w, http.StatusOK, requests)
}

// ------------------------------------------------------------- group posts

// CreateGroupPost handles POST /groups/{id}/posts (members only).
func (h *Handler) CreateGroupPost(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	post, err := h.Post.CreateGroupPost(userID, groupID, r.FormValue("content"), r.FormValue("privacy"))
	if err != nil {
		writeGroupPostError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, post)
}

// ListGroupPosts handles GET /groups/{id}/posts (members only).
func (h *Handler) ListGroupPosts(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	posts, err := h.Post.GroupPosts(userID, groupID)
	if err != nil {
		writeGroupPostError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, posts)
}

func writeGroupPostError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, postsvc.ErrNotGroupMember):
		http.Error(w, "only group members can view or create group posts", http.StatusForbidden)
	case errors.Is(err, postsvc.ErrInvalidPrivacy):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "could not process group post", http.StatusInternalServerError)
	}
}

// ---------------------------------------------------------------- events

// CreateEvent handles POST /groups/{id}/events (members only).
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	event, err := h.Events.Create(userID, groupID, r.FormValue("title"), r.FormValue("description"), parseEventDateTime(r.FormValue("date"), r.FormValue("time")))
	if err != nil {
		writeEventError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, event)
}

// ListEvents handles GET /groups/{id}/events (members only).
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	groupID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	events, err := h.Events.List(userID, groupID)
	if err != nil {
		writeEventError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, events)
}

// RespondEvent handles POST /events/{id}/response (group members only).
func (h *Handler) RespondEvent(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	eventID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid event id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	choice := strings.TrimSpace(r.FormValue("choice"))
	going, notGoing, err := h.Events.Respond(userID, eventID, choice)
	if err != nil {
		writeEventError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{
		"event_id":        eventID,
		"my_choice":       choice,
		"going_count":     going,
		"not_going_count": notGoing,
	})
}

// parseEventDateTime combines the HTML date/time inputs (date "2006-01-02",
// time "15:04") into a time.Time. The zero time signals "missing", which the
// service rejects.
func parseEventDateTime(date, clock string) time.Time {
	if date == "" || clock == "" {
		return time.Time{}
	}
	value, err := time.ParseInLocation("2006-01-02 15:04", date+" "+clock, time.Local)
	if err != nil {
		return time.Time{}
	}
	return value
}

// MyEventResponse handles GET /events/{id}/response: the caller's current
// going / not-going answer (null before they answer).
func (h *Handler) MyEventResponse(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	eventID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid event id", http.StatusBadRequest)
		return
	}
	response, err := h.Events.MyResponse(userID, eventID)
	if err != nil {
		writeEventError(w, err)
		return
	}
	var choice *string
	if response != nil {
		choice = &response.Choice
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"event_id": eventID, "my_choice": choice})
}

func writeEventError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, eventsvc.ErrInvalidTitle),
		errors.Is(err, eventsvc.ErrInvalidDescription),
		errors.Is(err, eventsvc.ErrInvalidDateTime),
		errors.Is(err, eventsvc.ErrInvalidChoice):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, eventsvc.ErrNotFound),
		errors.Is(err, repository.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, eventsvc.ErrNotGroupMember):
		http.Error(w, "only group members can do that", http.StatusForbidden)
	default:
		http.Error(w, "could not process event", http.StatusInternalServerError)
	}
}

// writeError maps service and repository errors to the project's flat-text
// http.Error responses with the same status-code vocabulary as the other
// handlers.
func writeError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, groupsvc.ErrInvalidTitle),
		errors.Is(err, groupsvc.ErrInvalidDescription),
		errors.Is(err, groupsvc.ErrSelfInvite),
		errors.Is(err, groupsvc.ErrSelfRequest):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, groupsvc.ErrNotFound),
		errors.Is(err, repository.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, groupsvc.ErrAlreadyMember),
		errors.Is(err, groupsvc.ErrInvitationExists),
		errors.Is(err, groupsvc.ErrRequestExists),
		errors.Is(err, repository.ErrExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, groupsvc.ErrNotGroupMember),
		errors.Is(err, groupsvc.ErrNotGroupCreator),
		errors.Is(err, groupsvc.ErrNotRecipient),
		errors.Is(err, repository.ErrNotOwner):
		http.Error(w, err.Error(), http.StatusForbidden)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}

func parseID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id < 1 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}
