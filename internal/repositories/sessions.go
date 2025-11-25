package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/google/uuid"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *models.Session) (*models.Session, error) {
	const q = `
INSERT INTO sessions (user_id, performed_at, notes, session_type)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, performed_at, notes, session_type`

	var created models.Session
	err := r.db.QueryRowContext(ctx, q, s.UserID, s.PerformedAt, s.Notes, s.SessionType).
		Scan(&created.ID, &created.UserID, &created.PerformedAt, &created.Notes, &created.SessionType)
	return &created, err
}

func (r *SessionRepository) AddSet(ctx context.Context, set *models.Set) (*models.Set, error) {
	const q = `
INSERT INTO sets (session_id, exercise_id, set_index, reps, weight_kg, rpe)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, session_id, exercise_id, set_index, reps, weight_kg, rpe`

	var out models.Set
	err := r.db.QueryRowContext(ctx, q, set.SessionID, set.ExerciseID, set.SetIndex, set.Reps, set.WeightKG, set.RPE).
		Scan(&out.ID, &out.SessionID, &out.ExerciseID, &out.SetIndex, &out.Reps, &out.WeightKG, &out.RPE)
	return &out, err
}

func (r *SessionRepository) ListWithSets(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	if userID == uuid.Nil {
		return nil, errors.New("userID cannot be nil")
	}

	const q = `
SELECT s.id, s.user_id, s.performed_at, s.notes, s.session_type,
       st.id, st.session_id, st.exercise_id, st.set_index, st.reps, st.weight_kg, st.rpe
FROM sessions s
LEFT JOIN sets st ON st.session_id = s.id
WHERE s.user_id = $1
ORDER BY s.performed_at DESC, st.set_index ASC`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make(map[string]*models.Session)
	for rows.Next() {
		var s models.Session
		var notes sql.NullString
		var sessionType sql.NullString
		var setID sql.NullString
		var setSessionID sql.NullString
		var exerciseID sql.NullString
		var setIndex sql.NullInt64
		var reps sql.NullInt64
		var weight sql.NullFloat64
		var rpe sql.NullInt64

		err = rows.Scan(
			&s.ID, &s.UserID, &s.PerformedAt, &notes, &sessionType,
			&setID, &setSessionID, exerciseID, &setIndex, &reps, &weight, &rpe,
		)
		if err != nil {
			return nil, err
		}

		if notes.Valid {
			s.Notes = &notes.String
		}
		if sessionType.Valid {
			s.SessionType = &sessionType.String
		}

		session, ok := sessions[s.ID.String()]
		if !ok {
			session = &models.Session{
				ID:          s.ID,
				UserID:      s.UserID,
				PerformedAt: s.PerformedAt,
				Notes:       s.Notes,
				SessionType: s.SessionType,
				Sets:        []models.Set{},
			}
			sessions[s.ID.String()] = session
		}

		if setID.Valid {
			// parse ID
			idParsed, err := uuid.Parse(setID.String)
			if err != nil {
				return nil, err
			}
			sid, err := uuid.Parse(setSessionID.String)
			if err != nil {
				return nil, err
			}
			eid, err := uuid.Parse(exerciseID.String)
			if err != nil {
				return nil, err
			}

			set := models.Set{
				ID:         idParsed,
				SessionID:  sid,
				ExerciseID: eid,
				SetIndex:   int(setIndex.Int64),
				Reps:       int(reps.Int64),
				WeightKG:   weight.Float64,
			}
			if rpe.Valid {
				value := int(rpe.Int64)
				set.RPE = &value
			}
			session.Sets = append(session.Sets, set)
		}
	}

	result := make([]*models.Session, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, sess)
	}
	return result, rows.Err()
}

func (r *SessionRepository) GetLastSet(ctx context.Context, userID uuid.UUID) (*models.Set, error) {
	const q = `
SELECT st.id, st.session_id, st.exercise_id, st.set_index, st.reps, st.weight_kg, st.rpe
FROM sets st
JOIN sessions s ON st.session_id = s.id
WHERE s.user_id = $1
ORDER BY st.created_at DESC NULLS LAST, st.set_index DESC
LIMIT 1`

	var set models.Set
	err := r.db.QueryRowContext(ctx, q, userID).
		Scan(&set.ID, &set.SessionID, &set.ExerciseID, &set.SetIndex, &set.Reps, &set.WeightKG, &set.RPE)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return &set, err
}

func (r *SessionRepository) GetLastSession(ctx context.Context, userID uuid.UUID) (*models.Session, error) {
	const q = `
SELECT id, user_id, performed_at, notes, session_type
FROM sessions
WHERE user_id = $1
ORDER BY performed_at DESC
LIMIT 1`

	var s models.Session
	var notes sql.NullString
	var sessionType sql.NullString
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&s.ID, &s.UserID, &s.PerformedAt, &notes, &sessionType)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if notes.Valid {
		s.Notes = &notes.String
	}
	if sessionType.Valid {
		s.SessionType = &sessionType.String
	}
	return &s, err
}

func (r *SessionRepository) SessionBelongsToUser(ctx context.Context, sessionID, userID uuid.UUID) (bool, error) {
	const q = `
SELECT 1
FROM sessions
WHERE id = $1 AND user_id = $2`

	var exists int
	err := r.db.QueryRowContext(ctx, q, sessionID, userID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
