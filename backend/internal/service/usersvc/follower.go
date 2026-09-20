package usersvc

// IsFollower reports whether viewerID has an accepted follow of targetID. The
// profile endpoint uses it to decide whether the full registration details
// (email, date of birth) belong in the response.
func (s *Service) IsFollower(viewerID, targetID int64) (bool, error) {
	if viewerID == 0 || viewerID == targetID {
		return false, nil
	}
	return s.users.IsFollowing(viewerID, targetID)
}
