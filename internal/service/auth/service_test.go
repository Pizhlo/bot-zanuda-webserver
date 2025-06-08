package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	type test struct {
		name string
		opts []AuthOption
		want *Service
		err  error
	}

	secretKey := []byte("secret")

	tests := []test{
		{
			name: "positive case",
			opts: []AuthOption{
				WithSecretKey(secretKey),
			},
			want: &Service{
				secretKey: secretKey,
			},
			err: nil,
		},
		{
			name: "error case: secret key is required",
			opts: []AuthOption{},
			err:  errors.New("secret key is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := New(tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
				assert.Nil(t, auth)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, auth)
			}
		})
	}
}

func createTestAuthService(t *testing.T, secretKey []byte) *Service {
	auth, err := New(WithSecretKey(secretKey))
	require.NoError(t, err)
	return auth
}
