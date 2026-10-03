package repository

import (
	"time"

	"sn-backend/internal/model"
)

const eventColumns = `
	e.id, e.group_id, e.creator_id, e.title, e.description, e.date_time, e.created_at,
	u.first_name, u.last_name`

func (r *Repository) CreateEvent(event *model.GroupEvent) error {
	result, err := r.db.Exec(
		`INSERT INTO group_events (group_id, creator_id, title, description, date_time)
		 VALUES (?, ?, ?, ?, ?)`,
		event.GroupID, event.CreatorID, event.Title, event.Description, event.DateTime,
	)
	if err != nil {
		return err
	}
	event.ID, err = result.LastInsertId()
	return err
}

// ListGroupEvents returns the events of a group (soonest first) with going /
// not-going counts and the viewer's own choice filled in.
func (r *Repository) ListGroupEvents(groupID, viewerID int64) ([]*model.EventListItem, error) {
	rows, err := r.db.Query(
		`SELECT `+eventColumns+`,
			COALESCE(SUM(CASE WHEN er.choice = ? THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN er.choice = ? THEN 1 ELSE 0 END), 0),
			COALESCE(MAX(CASE WHEN er.user_id = ? THEN er.choice END), '')
		 FROM group_events e
		 JOIN users u ON u.id = e.creator_id
		 LEFT JOIN event_responses er ON er.event_id = e.id
		 WHERE e.group_id = ?
		 GROUP BY e.id
		 ORDER BY e.date_time, e.id`,
		model.EventChoiceGoing, model.EventChoiceNotGoing, viewerID, groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*model.EventListItem, 0)
	for rows.Next() {
		// one Scan per row: event columns first, then the three aggregates
		item := new(model.EventListItem)
		if err := rows.Scan(
			&item.ID,
			&item.GroupID,
			&item.CreatorID,
			&item.Title,
			&item.Description,
			&item.DateTime,
			&item.CreatedAt,
			&item.CreatorFirstName,
			&item.CreatorLastName,
			&item.GoingCount,
			&item.NotGoingCount,
			&item.MyChoice,
		); err != nil {
			return nil, err
		}
		events = append(events, item)
	}
	return events, rows.Err()
}

// ListUpcomingEvents returns the next `limit` events, soonest first, across
// every group the viewer is a member of. The future check is done in Go so it
// does not depend on how the driver formats stored timestamps.
func (r *Repository) ListUpcomingEvents(viewerID int64, after time.Time, limit int) ([]*model.UpcomingEvent, error) {
	rows, err := r.db.Query(
		`SELECT e.id, e.group_id, e.creator_id, e.title, e.description, e.date_time, e.created_at, g.title
		 FROM group_events e
		 JOIN group_members gm ON gm.group_id = e.group_id AND gm.user_id = ?
		 JOIN groups g ON g.id = e.group_id
		 ORDER BY e.date_time, e.id`,
		viewerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]*model.UpcomingEvent, 0, limit)
	for rows.Next() && len(events) < limit {
		item := new(model.UpcomingEvent)
		if err := rows.Scan(
			&item.ID,
			&item.GroupID,
			&item.CreatorID,
			&item.Title,
			&item.Description,
			&item.DateTime,
			&item.CreatedAt,
			&item.GroupTitle,
		); err != nil {
			return nil, err
		}
		if item.DateTime.After(after) {
			events = append(events, item)
		}
	}
	return events, rows.Err()
}

// SetEventResponse inserts or replaces the user's response to an event. The
// UNIQUE(event_id, user_id) constraint guarantees one row per user + event.
func (r *Repository) SetEventResponse(eventID, userID int64, choice string) error {
	result, err := r.db.Exec(
		`INSERT INTO event_responses (event_id, user_id, choice) VALUES (?, ?, ?)
		 ON CONFLICT (event_id, user_id) DO UPDATE SET choice = excluded.choice`,
		eventID, userID, choice,
	)
	if err != nil {
		return err
	}
	// Distinguish "no row changed" from success; the INSERT arm always
	// changes a row, so RowsAffected is 0 only on driver anomalies.
	if count, err := result.RowsAffected(); err == nil && count == 0 {
		return ErrExists
	}
	return nil
}

// GetEventResponse returns the user's current response to an event,
// ErrNotFound when they have not responded yet.
func (r *Repository) DeleteEventResponse(eventID, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM event_responses WHERE event_id = ? AND user_id = ?`, eventID, userID)
	return err
}

func (r *Repository) GetEventResponse(eventID, userID int64) (*model.EventResponse, error) {
	response := new(model.EventResponse)
	err := r.QueryRow(
		`SELECT id, event_id, user_id, choice, created_at
		 FROM event_responses WHERE event_id = ? AND user_id = ?`,
		eventID, userID,
	).Scan(&response.ID, &response.EventID, &response.UserID, &response.Choice, &response.CreatedAt)
	if err != nil {
		return nil, notFound(err)
	}
	return response, nil
}

// EventResponseCounts returns going / not-going counts for one event.
func (r *Repository) EventResponseCounts(eventID int64) (going, notGoing int, err error) {
	err = r.QueryRow(
		`SELECT
			COALESCE(SUM(CASE WHEN choice = ? THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN choice = ? THEN 1 ELSE 0 END), 0)
		 FROM event_responses WHERE event_id = ?`,
		model.EventChoiceGoing, model.EventChoiceNotGoing, eventID,
	).Scan(&going, &notGoing)
	return going, notGoing, err
}
