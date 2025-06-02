package auth

import (
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckToken(t *testing.T) {
	type test struct {
		name  string
		token string
		want  jwt.MapClaims
		err   error
	}

	// payload: {"user_id":123,"expired":1875018533}
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsImV4cGlyZWQiOjE4NzUwMTg1MzN9.6-4hUO7BZAiODFldrNX8pHh0L0IsaK_iui0zCuhxcyI"

	// payload: {"user_id":123,"expired":1717252133}
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsImV4cGlyZWQiOjE3MTcyNTIxMzN9._bF-Vj10k_pgLaNVxTrftwPCKbdZEoicB3EOhoQCK_U"

	// payload: {"user_id":123,"expired":"1717252133"}
	expiredWithString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsImV4cGlyZWQiOiIxNzE3MjUyMTMzIn0.1AJLYAYrlYF6nnC3lzxoc5oYM4o66LQnEZS9fty-I7I"

	// payload: {"user_id":123}
	tokenWithoutExpiredKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjN9.PZLMJBT9OIVG2qgp9hQr685oVYFgRgWpcSPmNcw6y7M"

	tests := []test{
		{
			name:  "positive case",
			token: "Bearer " + token,
			want: jwt.MapClaims{
				"user_id": float64(123),
				"expired": float64(1875018533),
			},
			err: nil,
		},
		{
			name:  "error case: invalid token",
			token: "Bearer invalid",
			err:   errors.New("invalid token"),
		},
		{
			name:  "error case: expired",
			token: "Bearer " + expiredToken,
			err:   errors.New("token is expired"),
		},
		{
			name:  "error case: failed to convert interface to int64",
			token: "Bearer " + expiredWithString,
			err:   errors.New("failed to convert interface to int64"),
		},
		{
			name:  "error case: not found 'expired' key",
			token: "Bearer " + tokenWithoutExpiredKey,
			err:   errors.New("not found 'expired' key"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := createTestAuthService(t, []byte("secret"))
			token, err := auth.CheckToken(tt.token)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
				assert.Nil(t, token)
			} else {
				require.NoError(t, err)
				require.NotNil(t, token)
				assert.True(t, token.Valid)
				assert.Equal(t, "HS256", token.Method.Alg())
				assert.Equal(t, tt.want, token.Claims)
			}
		})
	}
}

func TestGetPayload(t *testing.T) {
	type test struct {
		name       string
		token      *jwt.Token
		wantClaims jwt.MapClaims
		wantOk     bool
	}

	validClaims := jwt.MapClaims{
		"user_id": float64(123),
		"expired": float64(1875018533),
	}

	tests := []test{
		{
			name: "positive case",
			token: &jwt.Token{
				Claims: jwt.MapClaims{
					"user_id": float64(123),
					"expired": float64(1875018533),
				},
			},
			wantClaims: validClaims,
			wantOk:     true,
		},
		{
			name: "error case: invalid claims type",
			token: &jwt.Token{
				Claims: &jwt.RegisteredClaims{}, // используем указатель на RegisteredClaims
			},
			wantClaims: nil,
			wantOk:     false,
		},
		{
			name:       "error case: nil token",
			token:      nil,
			wantClaims: nil,
			wantOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := createTestAuthService(t, []byte("secret"))
			claims, ok := auth.GetPayload(tt.token)
			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantClaims, claims)
		})
	}
}
