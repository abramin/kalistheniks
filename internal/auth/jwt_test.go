package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndParseToken(t *testing.T) {
	t.Run("round trip success", func(t *testing.T) {
		secret := "supersecret"
		userID := "user-123"

		token, err := GenerateToken(userID, secret)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		parsedUser, err := ParseToken(token, secret)
		require.NoError(t, err)
		require.Equal(t, userID, parsedUser)
	})

	t.Run("invalid signature", func(t *testing.T) {
		token, err := GenerateToken("user-456", "secret-a")
		require.NoError(t, err)

		parsedUser, err := ParseToken(token, "secret-b")
		require.Error(t, err)
		require.Empty(t, parsedUser)
	})

	t.Run("unexpected signing method", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.RegisteredClaims{
			Subject: "user-789",
		})
		signed, err := token.SignedString([]byte("secret"))
		require.NoError(t, err)

		parsedUser, err := ParseToken(signed, "secret")
		require.Error(t, err)
		require.Empty(t, parsedUser)
	})

	t.Run("expired token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   "user-expired",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		})
		signed, err := token.SignedString([]byte("secret"))
		require.NoError(t, err)

		parsedUser, err := ParseToken(signed, "secret")
		require.Error(t, err)
		require.Empty(t, parsedUser)
	})

	t.Run("malformed token", func(t *testing.T) {
		parsedUser, err := ParseToken("this.is.not.a.valid.token", "secret")
		require.Error(t, err)
		require.Empty(t, parsedUser)
	})

	t.Run("token contains correct claims", func(t *testing.T) {
		secret := "anothersecret"
		userID := "user-claims"

		tokenString, err := GenerateToken(userID, secret)
		require.NoError(t, err)
		token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		require.NoError(t, err)

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		require.True(t, ok)
		require.Equal(t, userID, claims.Subject)
		require.Equal(t, "kalistheniks-api", claims.Issuer)
		require.Contains(t, claims.Audience, "kalistheniks-users")
		require.NotEmpty(t, claims.ID)
	})
}
