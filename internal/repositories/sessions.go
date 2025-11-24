package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/alexanderramin/kalistheniks/internal/models"
)

type SessionRepository interface {
	Create(ctx context.Context, s models.Session) (models.Session, error)
	AddSet(ctx context.Context, set models.Set) (models.Set, error)
	ListWithSets(ctx context.Context, userID string) ([]models.Session, error)
	GetLastSet(ctx context.Context, userID string) (models.Set, error)
	GetLastSession(ctx context.Context, userID string) (models.Session, error)
}

type sessionRepo struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(ctx context.Context, s models.Session) (models.Session, error) {
	const q = `
INSERT INTO sessions (user_id, performed_at, notes, session_type)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, performed_at, notes, session_type`

	var created models.Session
	err := r.db.QueryRowContext(ctx, q, s.UserID, s.PerformedAt, s.Notes, s.SessionType).
		Scan(&created.ID, &created.UserID, &created.PerformedAt, &created.Notes, &created.SessionType)
	return created, err
}

func (r *sessionRepo) AddSet(ctx context.Context, set models.Set) (models.Set, error) {
	const q = `
INSERT INTO sets (session_id, exercise_id, set_index, reps, weight_kg, rpe)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, session_id, exercise_id, set_index, reps, weight_kg, rpe`

	var out models.Set
	err := r.db.QueryRowContext(ctx, q, set.SessionID, set.ExerciseID, set.SetIndex, set.Reps, set.WeightKG, set.RPE).
		Scan(&out.ID, &out.SessionID, &out.ExerciseID, &out.SetIndex, &out.Reps, &out.WeightKG, &out.RPE)
	return out, err
}

func (r *sessionRepo) ListWithSets(ctx context.Context, userID string) ([]models.Session, error) {
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
			&setID, &setSessionID, &exerciseID, &setIndex, &reps, &weight, &rpe,
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

		session, ok := sessions[s.ID]
		if !ok {
			session = &models.Session{
				ID:          s.ID,
				UserID:      s.UserID,
				PerformedAt: s.PerformedAt,
				Notes:       s.Notes,
				SessionType: s.SessionType,
				Sets:        []models.Set{},
			}
			sessions[s.ID] = session
		}

		if setID.Valid {
			set := models.Set{
				ID:         setID.String,
				SessionID:  setSessionID.String,
				ExerciseID: exerciseID.String,
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

	result := make([]models.Session, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, *sess)
	}
	return result, rows.Err()
}

func (r *sessionRepo) GetLastSet(ctx context.Context, userID string) (models.Set, error) {
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
		return models.Set{}, err
	}
	return set, err
}

func (r *sessionRepo) GetLastSession(ctx context.Context, userID string) (models.Session, error) {
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
		return models.Session{}, err
	}
	if notes.Valid {
		s.Notes = &notes.String
	}
	if sessionType.Valid {
		s.SessionType = &sessionType.String
	}
	return s, err
}
