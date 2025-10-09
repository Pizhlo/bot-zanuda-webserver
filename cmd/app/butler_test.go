package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewButler(t *testing.T) {
	t.Parallel()

	setBuildVars(t, "v1.0.0", "2024-01-15", "abc123def456")

	butler := NewButler()
	require.NotNil(t, butler)

	assert.Equal(t, "v1.0.0", butler.BuildInfo.Version)
	assert.Equal(t, "2024-01-15", butler.BuildInfo.BuildDate)
	assert.Equal(t, "abc123def456", butler.BuildInfo.GitCommit)
}

func TestStart(t *testing.T) {
	t.Parallel()

	ch := make(chan bool)

	start := func() error {
		ch <- true
		return nil
	}

	butler := NewButler()
	butler.start(start)

	select {
	case <-ch:
		assert.True(t, true)
	case <-time.After(1 * time.Second):
		assert.Fail(t, "start function did not return")
	}
}

type mockStopper struct {
	ch  chan bool
	err error
}

func (m *mockStopper) Stop(_ context.Context) error {
	m.ch <- true
	return m.err
}

func TestStop(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	type testCase struct {
		name    string
		stopper stopper
		err     error
	}

	tests := []testCase{
		{name: "positive case", stopper: &mockStopper{ch: make(chan bool), err: nil}, err: nil},
		{name: "negative case", stopper: &mockStopper{ch: make(chan bool), err: errors.New("test error")}, err: errors.New("test error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			go func() {
				select {
				case <-tt.stopper.(*mockStopper).ch:
					assert.True(t, true)
				case <-time.After(1 * time.Second):
					assert.Fail(t, "stop function did not return")
				}
			}()

			butler := NewButler()
			butler.stop(ctx, tt.stopper)
		})
	}
}
