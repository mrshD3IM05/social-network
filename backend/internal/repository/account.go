package repository

import "sn-backend/internal/model"

// SetUserPrivacy flips the profile between public and private.
func (r *Repository) SetUserPrivacy(userID int64, private bool) error {
	result, err := r.db.Exec(`UPDATE users SET private = ? WHERE id = ?`, boolToInt(private), userID)
	if err != nil {
		return err
	}
	if count, err := result.RowsAffected(); err != nil {
		return err
	} else if count == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateProfileFields writes the editable profile fields of one account.
func (r *Repository) UpdateProfileFields(user *model.User) error {
	result, err := r.db.Exec(
		`UPDATE users SET first_name = ?, last_name = ?, nickname = ?, about_me = ?, private = ? WHERE id = ?`,
		user.FirstName, user.LastName, user.Nickname, user.AboutMe, boolToInt(user.Private), user.ID,
	)
	if err != nil {
		return err
	}
	if count, err := result.RowsAffected(); err != nil {
		return err
	} else if count == 0 {
		return ErrNotFound
	}
	return nil
}

// NicknameTaken reports whether another account already uses this nickname.
func (r *Repository) NicknameTaken(nickname string, exceptUserID int64) (bool, error) {
	if nickname == "" {
		return false, nil
	}
	var taken int
	err := r.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE nickname = ? AND id <> ?)`,
		nickname, exceptUserID,
	).Scan(&taken)
	return taken == 1, err
}

// DeleteExpiredSessions reaps session rows whose expiry has passed. Without it
// dead rows only disappear when someone happens to present the stale cookie.
func (r *Repository) DeleteExpiredSessions() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
