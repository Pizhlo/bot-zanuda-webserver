package server

import (
	"fmt"
	"testing"
	v0 "webserver/internal/server/api/v0"

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

	tests := []test{
		{
			name: "positive case",
			params: struct {
				addr string
				h0   handler
			}{
				addr: "addr",
				h0:   v0.New(nil, nil),
			},
			result: struct {
				config *Config
				err    error
			}{
				config: &Config{
					Address:   "addr",
					HandlerV0: v0.New(nil, nil),
				},
			},
		},
		{
			name: "empty addr",
			params: struct {
				addr string
				h0   handler
			}{
				h0: v0.New(nil, nil),
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
