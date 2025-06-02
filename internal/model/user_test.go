package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserValidate(t *testing.T) {
	type test struct {
		name string
		user User
		err  error
	}

	tests := []test{
		{
			name: "positive case",
			user: User{
				ID:       345,
				TgID:     123,
				Username: "username",
				PersonalSpace: &Space{
					ID:       uuid.New(),
					Name:     "space",
					Created:  time.Now(),
					Creator:  123,
					Personal: true,
				},
				Timezone: "timezone",
			},
		},
		{
			name: "id not filled",
			user: User{
				TgID:     123,
				Username: "username",
				PersonalSpace: &Space{
					ID:       uuid.New(),
					Name:     "space",
					Created:  time.Now(),
					Creator:  123,
					Personal: true,
				},
				Timezone: "timezone",
			},
			err: ErrFieldIDNotFilled,
		},
		{
			name: "empty tg id",
			user: User{
				ID:       345,
				Username: "username",
				PersonalSpace: &Space{
					ID:       uuid.New(),
					Name:     "space",
					Created:  time.Now(),
					Creator:  123,
					Personal: true,
				},
				Timezone: "timezone",
			},
			err: ErrTgIDNotFilled,
		},
		{
			name: "username is empty",
			user: User{
				ID:   345,
				TgID: 123,
				PersonalSpace: &Space{
					ID:       uuid.New(),
					Name:     "space",
					Created:  time.Now(),
					Creator:  123,
					Personal: true,
				},
				Timezone: "timezone",
			},
			err: ErrUsernameNotFilled,
		},
		{
			name: "personal space is nil",
			user: User{
				ID:       345,
				TgID:     123,
				Username: "username",
				Timezone: "timezone",
			},
			err: ErrSpaceIsNil,
		},
		{
			name: "timezone empty",
			user: User{
				ID:       345,
				TgID:     123,
				Username: "username",
				PersonalSpace: &Space{
					ID:       uuid.New(),
					Name:     "space",
					Created:  time.Now(),
					Creator:  123,
					Personal: true,
				},
			},
			err: ErrTimezoneNotFilled,
		},
		{
			name: "invalid space",
			user: User{
				ID:       345,
				TgID:     123,
				Username: "username",
				PersonalSpace: &Space{
					ID:       uuid.New(),
					Name:     "space",
					Created:  time.Now(),
					Personal: true,
				},
				Timezone: "timezone",
			},
			err: ErrFieldCreatorNotFilled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if tt.err != nil {
				assert.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
