package server

import (
	"fmt"
	"testing"
	"webserver/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	type test struct {
		name   string
		params struct {
			addr string
			h0   handler
		}

		result struct {
			config *Config
			err    error
		}
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := mocks.NewMockhandler(ctrl)

	tests := []test{
		{
			name: "positive case",
			params: struct {
				addr string
				h0   handler
			}{
				addr: "addr",
				h0:   h,
			},
			result: struct {
				config *Config
				err    error
			}{
				config: &Config{
					Address:   "addr",
					HandlerV0: h,
				},
			},
		},
		{
			name: "empty addr",
			params: struct {
				addr string
				h0   handler
			}{
				h0: h,
			},
			result: struct {
				config *Config
				err    error
			}{
				err: fmt.Errorf("address is empty"),
			},
		},
		{
			name: "handler is nil",
			params: struct {
				addr string
				h0   handler
			}{
				addr: "addr",
			},
			result: struct {
				config *Config
				err    error
			}{
				err: fmt.Errorf("handler is nil"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := NewConfig(tt.params.addr, tt.params.h0)
			if tt.result.err != nil {
				assert.EqualError(t, err, tt.result.err.Error())
			} else {
				assert.Equal(t, tt.result.config, actual)
				assert.NoError(t, err)
			}
		})
	}
}
