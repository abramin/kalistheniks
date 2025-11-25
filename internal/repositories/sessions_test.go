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
	s.truncateSessions()
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

func TestSessionRepository_ListWithSets(t *testing.T) {
	t.Skip("TODO: implement session repository list with sets test")
	_ = require.New(t)
}

func TestSessionRepository_GetLastSet(t *testing.T) {
	t.Skip("TODO: implement session repository get last set test")
	_ = require.New(t)
}

func TestSessionRepository_GetLastSession(t *testing.T) {
	t.Skip("TODO: implement session repository get last session test")
	_ = require.New(t)
}

func TestSessionRepository_SessionBelongsToUser(t *testing.T) {
	t.Skip("TODO: implement session repository session belongs to user test")
	_ = require.New(t)
}

func ptrToString(s string) *string {
	return &s
}

func (s *SessionRepositorySuite) truncateSessions() {
	s.T().Helper()
	_, err := testDB.Exec("TRUNCATE TABLE sets RESTART IDENTITY CASCADE; TRUNCATE TABLE sessions RESTART IDENTITY CASCADE")
	require.NoError(s.T(), err)
}
