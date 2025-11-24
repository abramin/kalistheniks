package repositories

import (
	"context"
	"testing"

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
}

func (s *SessionRepositorySuite) SetupTest() {
	truncateSessions(s.T())
	s.sessionRepo = NewSessionRepository(testDB)

}

func (s *SessionRepositorySuite) TestCreateSessionSuccess() {

	session := models.Session{
		UserID:      s.user.ID,
		SessionType: ptrToString("workout"),
	}

	created, err := s.sessionRepo.Create(s.ctx, session)
	s.Require().NoError(err)
	s.Require().Equal(s.user.ID, created.UserID)
	s.Require().Equal("workout", *created.SessionType)
	s.Require().NotEmpty(created.ID)
}

func (s *SessionRepositorySuite) TestCreateSessionInvalidUser() {
	ctx := context.Background()
	session := models.Session{
		UserID:      uuid.NewString(),
		SessionType: ptrToString("workout"),
	}

	_, err := s.sessionRepo.Create(ctx, session)
	s.Require().Error(err)
}

func TestSessionRepository_AddSet(t *testing.T) {
	t.Run("adds set successfully", func(t *testing.T) {

	})

	t.Run("invalid session ID raises error", func(t *testing.T) {
		t.Skip("TODO: implement session repository add set invalid id test")
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

func truncateSessions(t *testing.T) {

	_, err := testDB.Exec("TRUNCATE TABLE sessions RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}
