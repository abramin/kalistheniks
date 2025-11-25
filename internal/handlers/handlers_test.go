package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexanderramin/kalistheniks/internal/handlers/mocks"
	"github.com/alexanderramin/kalistheniks/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

//go:generate mockgen -source=./contracts/contracts.go -destination=./mocks/mocks.go -package=mocks AuthService,SessionService,PlanService
type HandlerSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	authMock    *mocks.MockAuthService
	sessionMock *mocks.MockSessionService
	planMock    *mocks.MockPlanService
	handler     http.Handler
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

func (s *HandlerSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.authMock = mocks.NewMockAuthService(s.ctrl)
	s.sessionMock = mocks.NewMockSessionService(s.ctrl)
	s.planMock = mocks.NewMockPlanService(s.ctrl)

	app := &App{
		AuthService:    s.authMock,
		SessionService: s.sessionMock,
		PlanService:    s.planMock,
	}
	s.handler = Router(app)
}

func (s *HandlerSuite) TearDownTest() {
	if s.ctrl != nil {
		s.ctrl.Finish()
	}
}

func (s *HandlerSuite) TestHealth() {
	resp := s.doRequest(http.MethodGet, "/health", nil, "")
	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *HandlerSuite) TestSignup() {
	s.Run("success", func() {
		user := &models.User{ID: uuid.New()}
		s.authMock.EXPECT().Signup(gomock.Any(), "user@example.com", "Password123").Return(user, "token", nil)

		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString(`{"email":"user@example.com","password":"Password123"}`), "")
		s.Equal(http.StatusCreated, resp.StatusCode)
	})

	s.Run("bad input", func() {
		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString("not-json"), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("invalid email format", func() {
		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString(`{"email":"notanemail","password":"Password123"}`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("weak password", func() {
		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString(`{"email":"user@example.com","password":"weak"}`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("password too short", func() {
		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString(`{"email":"user@example.com","password":"Pass1"}`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("malformed JSON", func() {
		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString(`{"email":"user@example.com","password":`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("empty email", func() {
		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString(`{"email":"","password":"Password123"}`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("empty body", func() {
		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString(`{}`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("unknown fields", func() {
		resp := s.doRequest(http.MethodPost, "/signup", bytes.NewBufferString(`{"email":"user@example.com","password":"Password123","extra":"field"}`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (s *HandlerSuite) TestLogin() {
	s.Run("success", func() {
		user := &models.User{ID: uuid.New()}
		s.authMock.EXPECT().Login(gomock.Any(), "login@example.com", "password").Return(user, "token", nil)

		resp := s.doRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"login@example.com","password":"password"}`), "")
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.Run("invalid credentials", func() {
		s.authMock.EXPECT().Login(gomock.Any(), "login@example.com", "wrong").Return(nil, "", errors.New("invalid credentials"))

		resp := s.doRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"login@example.com","password":"wrong"}`), "")
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("empty email", func() {
		resp := s.doRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"","password":"password"}`), "")
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("empty password", func() {
		resp := s.doRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"login@example.com","password":""}`), "")
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("malformed JSON", func() {
		resp := s.doRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"login@example.com"`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("unknown fields", func() {
		resp := s.doRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"login@example.com","password":"password","extra":"field"}`), "")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func (s *HandlerSuite) TestSessionEndpoints() {
	userID := uuid.New()

	s.Run("create session unauthorized", func() {
		resp := s.doRequest(http.MethodPost, "/sessions", bytes.NewBufferString(`{}`), "")
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("create session success", func() {
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)
		created := &models.Session{ID: uuid.New(), UserID: userID}
		s.sessionMock.EXPECT().CreateSession(gomock.Any(), userID, gomock.Nil(), gomock.Nil(), gomock.Nil()).Return(created, nil)

		resp := s.doRequest(http.MethodPost, "/sessions", bytes.NewBufferString(`{}`), "goodtoken")
		s.Equal(http.StatusCreated, resp.StatusCode)
	})

	s.Run("create set invalid session id", func() {
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)

		resp := s.doRequest(http.MethodPost, "/sessions/not-a-uuid/sets", bytes.NewBufferString(`{}`), "goodtoken")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("create set success", func() {
		sessionID := uuid.New()
		exerciseID := uuid.New()
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)
		s.sessionMock.EXPECT().AddSet(gomock.Any(), userID, sessionID, exerciseID, 0, 8, 20.0, gomock.Nil()).Return(&models.Set{ID: uuid.New(), SessionID: sessionID, ExerciseID: exerciseID}, nil)

		body := map[string]any{"exercise_id": exerciseID.String(), "set_index": 0, "reps": 8, "weight_kg": 20.0}
		payload, _ := json.Marshal(body)
		resp := s.doRequest(http.MethodPost, "/sessions/"+sessionID.String()+"/sets", bytes.NewBuffer(payload), "goodtoken")
		s.Equal(http.StatusCreated, resp.StatusCode)
	})

	s.Run("create set with negative reps", func() {
		sessionID := uuid.New()
		exerciseID := uuid.New()
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)

		body := map[string]any{"exercise_id": exerciseID.String(), "set_index": 0, "reps": -5, "weight_kg": 20.0}
		payload, _ := json.Marshal(body)
		resp := s.doRequest(http.MethodPost, "/sessions/"+sessionID.String()+"/sets", bytes.NewBuffer(payload), "goodtoken")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("create set with negative weight", func() {
		sessionID := uuid.New()
		exerciseID := uuid.New()
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)

		body := map[string]any{"exercise_id": exerciseID.String(), "set_index": 0, "reps": 8, "weight_kg": -10.0}
		payload, _ := json.Marshal(body)
		resp := s.doRequest(http.MethodPost, "/sessions/"+sessionID.String()+"/sets", bytes.NewBuffer(payload), "goodtoken")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("create set with invalid RPE", func() {
		sessionID := uuid.New()
		exerciseID := uuid.New()
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)

		rpe := 15 // RPE should be 1-10
		body := map[string]any{"exercise_id": exerciseID.String(), "set_index": 0, "reps": 8, "weight_kg": 20.0, "rpe": rpe}
		payload, _ := json.Marshal(body)
		resp := s.doRequest(http.MethodPost, "/sessions/"+sessionID.String()+"/sets", bytes.NewBuffer(payload), "goodtoken")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("create set with invalid exercise ID format", func() {
		sessionID := uuid.New()
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)

		body := map[string]any{"exercise_id": "not-a-uuid", "set_index": 0, "reps": 8, "weight_kg": 20.0}
		payload, _ := json.Marshal(body)
		resp := s.doRequest(http.MethodPost, "/sessions/"+sessionID.String()+"/sets", bytes.NewBuffer(payload), "goodtoken")
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.Run("list sessions unauthorized", func() {
		resp := s.doRequest(http.MethodGet, "/sessions", nil, "")
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("list sessions success", func() {
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)
		s.sessionMock.EXPECT().ListSessions(gomock.Any(), userID).Return([]*models.Session{}, nil)

		resp := s.doRequest(http.MethodGet, "/sessions", nil, "goodtoken")
		s.Equal(http.StatusOK, resp.StatusCode)
	})
}

func (s *HandlerSuite) TestPlanEndpoint() {
	userID := uuid.New()

	s.Run("unauthorized", func() {
		resp := s.doRequest(http.MethodGet, "/plan/next", nil, "")
		s.Equal(http.StatusUnauthorized, resp.StatusCode)
	})

	s.Run("success", func() {
		s.authMock.EXPECT().VerifyToken(gomock.Any(), "goodtoken").Return(userID.String(), nil)
		s.planMock.EXPECT().NextSuggestion(gomock.Any(), userID).Return(&models.PlanSuggestion{ExerciseID: uuid.New(), WeightKG: 20, Reps: 8}, nil)

		resp := s.doRequest(http.MethodGet, "/plan/next", nil, "goodtoken")
		s.Equal(http.StatusOK, resp.StatusCode)
	})
}

// helpers

func (s *HandlerSuite) doRequest(method, path string, body *bytes.Buffer, token string) *http.Response {
	var reader io.Reader
	if body != nil {
		reader = body
	}
	req, _ := http.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	s.handler.ServeHTTP(rec, req)
	return rec.Result()
}
