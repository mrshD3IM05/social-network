Dependency Injection
expalin architicture li rani khadam biha 
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request)
toUserID, err := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
