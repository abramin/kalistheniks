package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SessionRepositorySuite struct {
	suite.Suite
	sessionRepo *SessionRepository
	userRepo    *UserRepository
	ctx         context.Context
	user        *models.User
	exerciseID  *uuid.UUID
}

func TestSessionRepositorySuite(t *testing.T) {
	suite.Run(t, new(SessionRepositorySuite))
}

func (s *SessionRepositorySuite) SetupSuite() {
	s.ctx = context.Background()
	s.userRepo = NewUserRepository(testDB)
	var err error
	s.user, err = s.userRepo.Create(s.ctx, "session-user@example.com", "hash")
	s.Require().NoError(err)
	err = testDB.QueryRowContext(s.ctx, `INSERT INTO exercises (name) VALUES ('push-ups') RETURNING id`).Scan(&s.exerciseID)
	s.Require().NoError(err)
}

func (s *SessionRepositorySuite) SetupTest() {
	s.sessionRepo = NewSessionRepository(testDB)
}

func (s *SessionRepositorySuite) TestCreateSessionSuccess() {
	session := &models.Session{
		PerformedAt: time.Now().UTC(),
		Notes:       ptrToString("Test notes"),
		UserID:      s.user.ID,
		SessionType: ptrToString("workout"),
	}

	created, err := s.sessionRepo.Create(s.ctx, session)
	s.Require().NoError(err)
	s.Require().Equal(s.user.ID, created.UserID)
	s.Require().Equal("workout", *created.SessionType)
	s.Require().NotEmpty(created.ID)
}

func (s *SessionRepositorySuite) TestCreateSessionNoneExistentUser() {
	ctx := context.Background()
	otherID := uuid.New()
	session := &models.Session{
		UserID:      &otherID,
		SessionType: ptrToString("workout"),
	}

	_, err := s.sessionRepo.Create(ctx, session)
	s.Require().Error(err)
}

func (s *SessionRepositorySuite) TestSessionRepository_AddSet() {
	s.T().Run("adds set successfully", func(t *testing.T) {
		session := &models.Session{
			PerformedAt: time.Now().UTC(),
			Notes:       ptrToString("Test notes"),
			UserID:      s.user.ID,
			SessionType: ptrToString("workout"),
		}
		createdSession, err := s.sessionRepo.Create(context.Background(), session)
		require.NoError(t, err)
		rpe := 8
		set := &models.Set{
			SessionID:  createdSession.ID,
			ExerciseID: s.exerciseID,
			SetIndex:   0,
			Reps:       10,
			WeightKG:   0.0,
			RPE:        &rpe,
		}

		addedSet, err := s.sessionRepo.AddSet(context.Background(), set)
		require.NoError(t, err)
		require.Equal(t, createdSession.ID, addedSet.SessionID)
		require.Equal(t, s.exerciseID, addedSet.ExerciseID)
		require.Equal(t, 0, addedSet.SetIndex)
		require.Equal(t, 10, addedSet.Reps)
		require.Equal(t, 0.0, addedSet.WeightKG)
		require.NotNil(t, addedSet.RPE)
		require.Equal(t, 8, *addedSet.RPE)
	})

	s.T().Run("invalid session ID raises error", func(t *testing.T) {
		otherID := uuid.New()
		set := &models.Set{
			SessionID:  &otherID,
			ExerciseID: s.exerciseID,
			SetIndex:   0,
			Reps:       10,
			WeightKG:   0.0,
		}

		_, err := s.sessionRepo.AddSet(context.Background(), set)
		require.Error(t, err)
	})
}

func (s *SessionRepositorySuite) TestSessionRepository_ListWithSets() {
	s.T().Run("lists sessions with sets successfully", func(t *testing.T) {
		s.truncateSessions()
		// Create a session
		session := &models.Session{
			PerformedAt: time.Now().UTC(),
			Notes:       ptrToString("Test notes"),
			UserID:      s.user.ID,
			SessionType: ptrToString("workout"),
		}
		createdSession, err := s.sessionRepo.Create(context.Background(), session)
		require.NoError(t, err)

		// Add a set to the session
		set := &models.Set{
			SessionID:  createdSession.ID,
			ExerciseID: s.exerciseID,
			SetIndex:   0,
			Reps:       10,
			WeightKG:   0.0,
		}
		_, err = s.sessionRepo.AddSet(context.Background(), set)
		require.NoError(t, err)

		// List sessions with sets
		sessions, err := s.sessionRepo.ListWithSets(context.Background(), s.user.ID)
		require.NoError(t, err)
		require.Len(t, sessions, 1)
		require.Equal(t, createdSession.ID, sessions[0].ID)
		require.Len(t, sessions[0].Sets, 1)
		require.Equal(t, 10, sessions[0].Sets[0].Reps)
	})

	s.T().Run("no sessions returns empty list", func(t *testing.T) {
		s.truncateSessions()
		sessions, err := s.sessionRepo.ListWithSets(context.Background(), s.user.ID)
		require.NoError(t, err)
		require.Len(t, sessions, 0)
	})

	s.T().Run("invalid user ID returns empty list", func(t *testing.T) {
		s.truncateSessions()
		otherID := uuid.New()
		sessions, err := s.sessionRepo.ListWithSets(context.Background(), &otherID)
		require.NoError(t, err)
		require.Len(t, sessions, 0)
	})

	s.T().Run("nil user ID returns error", func(t *testing.T) {
		s.truncateSessions()
		_, err := s.sessionRepo.ListWithSets(context.Background(), nil)
		require.Error(t, err)
	})
}

func (s *SessionRepositorySuite) TestSessionRepository_GetLastSet() {
	s.T().Run("gets last set successfully", func(t *testing.T) {
		s.truncateSessions()
		session := &models.Session{
			PerformedAt: time.Now().UTC(),
			Notes:       ptrToString("Test notes"),
			UserID:      s.user.ID,
			SessionType: ptrToString("workout"),
		}
		createdSession, err := s.sessionRepo.Create(context.Background(), session)
		require.NoError(t, err)

		// Add two sets to the session
		set1 := &models.Set{
			SessionID:  createdSession.ID,
			ExerciseID: s.exerciseID,
			SetIndex:   0,
			Reps:       10,
			WeightKG:   0.0,
		}
		_, err = s.sessionRepo.AddSet(context.Background(), set1)
		require.NoError(t, err)

		set2 := &models.Set{
			SessionID:  createdSession.ID,
			ExerciseID: s.exerciseID,
			SetIndex:   1,
			Reps:       8,
			WeightKG:   0.0,
		}
		_, err = s.sessionRepo.AddSet(context.Background(), set2)
		require.NoError(t, err)

		// Get last set
		lastSet, err := s.sessionRepo.GetLastSet(context.Background(), s.user.ID)
		require.NoError(t, err)
		require.Equal(t, 8, lastSet.Reps)
		require.Equal(t, 1, lastSet.SetIndex)
	})

	s.T().Run("no sets returns nil", func(t *testing.T) {
		s.truncateSessions()
		_, err := s.sessionRepo.GetLastSet(context.Background(), s.user.ID)
		require.Error(t, err)
	})

	s.T().Run("nil user ID returns error", func(t *testing.T) {
		s.truncateSessions()
		_, err := s.sessionRepo.GetLastSet(context.Background(), nil)
		require.Error(t, err)
	})
}

func (s *SessionRepositorySuite) TestSessionRepository_GetLastSession() {
	s.T().Run("gets last session successfully", func(t *testing.T) {
		s.truncateSessions()
		// Create two sessions
		session1 := &models.Session{
			PerformedAt: time.Now().Add(-2 * time.Hour).UTC(),
			Notes:       ptrToString("First session"),
			UserID:      s.user.ID,
			SessionType: ptrToString("workout"),
		}
		_, err := s.sessionRepo.Create(context.Background(), session1)
		require.NoError(t, err)

		session2 := &models.Session{
			PerformedAt: time.Now().Add(-1 * time.Hour).UTC(),
			Notes:       ptrToString("Second session"),
			UserID:      s.user.ID,
			SessionType: ptrToString("workout"),
		}
		createdSession2, err := s.sessionRepo.Create(context.Background(), session2)
		require.NoError(t, err)

		// Get last session
		lastSession, err := s.sessionRepo.GetLastSession(context.Background(), s.user.ID)
		require.NoError(t, err)
		require.Equal(t, createdSession2.ID, lastSession.ID)
	})

	s.T().Run("no sessions returns nil", func(t *testing.T) {
		s.truncateSessions()
		_, err := s.sessionRepo.GetLastSession(context.Background(), s.user.ID)
		require.Error(t, err)
	})

	s.T().Run("nil user ID returns error", func(t *testing.T) {
		s.truncateSessions()
		_, err := s.sessionRepo.GetLastSession(context.Background(), nil)
		require.Error(t, err)
	})
}

func (s *SessionRepositorySuite) TestSessionRepository_SessionBelongsToUser() {
	s.T().Run("session belongs to user", func(t *testing.T) {
		session := &models.Session{
			PerformedAt: time.Now().UTC(),
			Notes:       ptrToString("Test notes"),
			UserID:      s.user.ID,
			SessionType: ptrToString("workout"),
		}
		createdSession, err := s.sessionRepo.Create(context.Background(), session)
		require.NoError(t, err)

		belongs, err := s.sessionRepo.SessionBelongsToUser(context.Background(), createdSession.ID, s.user.ID)
		require.NoError(t, err)
		require.True(t, belongs)
	})

	s.T().Run("session does not belong to user", func(t *testing.T) {
		otherID := uuid.New()
		belongs, err := s.sessionRepo.SessionBelongsToUser(context.Background(), &otherID, s.user.ID)
		require.NoError(t, err)
		require.False(t, belongs)
	})

	s.T().Run("	invalid session ID returns false", func(t *testing.T) {
		otherID := uuid.New()
		belongs, err := s.sessionRepo.SessionBelongsToUser(context.Background(), &otherID, s.user.ID)
		require.NoError(t, err)
		require.False(t, belongs)
	})
}

func ptrToString(s string) *string {
	return &s
}

func (s *SessionRepositorySuite) truncateSessions() {
	s.T().Helper()
	_, err := testDB.Exec("TRUNCATE TABLE sets RESTART IDENTITY CASCADE; TRUNCATE TABLE sessions RESTART IDENTITY CASCADE")
	require.NoError(s.T(), err)
}
