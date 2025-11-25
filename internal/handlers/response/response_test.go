package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "success response",
			status:         http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"message": "success"},
		},
		{
			name:           "created response",
			status:         http.StatusCreated,
			data:           map[string]interface{}{"id": "123", "name": "test"},
			expectedStatus: http.StatusCreated,
			expectedBody:   map[string]interface{}{"id": "123", "name": "test"},
		},
		{
			name:           "empty object",
			status:         http.StatusOK,
			data:           map[string]string{},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{},
		},
		{
			name:           "array response",
			status:         http.StatusOK,
			data:           []string{"item1", "item2"},
			expectedStatus: http.StatusOK,
			expectedBody:   nil, // We'll check this separately
		},
		{
			name:           "nil data",
			status:         http.StatusOK,
			data:           nil,
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			JSON(rec, tt.status, tt.data)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			// Verify JSON can be decoded
			if tt.data != nil {
				var result interface{}
				err := json.NewDecoder(rec.Body).Decode(&result)
				require.NoError(t, err)

				if tt.expectedBody != nil {
					resultMap, ok := result.(map[string]interface{})
					require.True(t, ok)
					for key, expectedVal := range tt.expectedBody {
						assert.Equal(t, expectedVal, resultMap[key])
					}
				}
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		message        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "bad request",
			status:         http.StatusBadRequest,
			message:        "invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid input",
		},
		{
			name:           "unauthorized",
			status:         http.StatusUnauthorized,
			message:        "unauthorized",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "unauthorized",
		},
		{
			name:           "not found",
			status:         http.StatusNotFound,
			message:        "resource not found",
			expectedStatus: http.StatusNotFound,
			expectedError:  "resource not found",
		},
		{
			name:           "internal server error",
			status:         http.StatusInternalServerError,
			message:        "internal error",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal error",
		},
		{
			name:           "empty message",
			status:         http.StatusBadRequest,
			message:        "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "",
		},
		{
			name:           "entity too large",
			status:         http.StatusRequestEntityTooLarge,
			message:        "request body too large",
			expectedStatus: http.StatusRequestEntityTooLarge,
			expectedError:  "request body too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Error(rec, tt.status, tt.message)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			// Verify error response structure
			var errorResp map[string]string
			err := json.NewDecoder(rec.Body).Decode(&errorResp)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedError, errorResp["error"])
		})
	}
}

func TestJSONWithComplexData(t *testing.T) {
	type User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	user := User{
		ID:    "123",
		Email: "test@example.com",
		Age:   30,
	}

	rec := httptest.NewRecorder()
	JSON(rec, http.StatusOK, user)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var result User
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestJSONWithNestedData(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id":    "123",
			"email": "test@example.com",
		},
		"token": "abc123",
		"metadata": map[string]interface{}{
			"created_at": "2024-01-01",
			"expires_in": 3600,
		},
	}

	rec := httptest.NewRecorder()
	JSON(rec, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, rec.Code)

	var result map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	assert.NotNil(t, result["user"])
	assert.NotNil(t, result["token"])
	assert.NotNil(t, result["metadata"])
}
