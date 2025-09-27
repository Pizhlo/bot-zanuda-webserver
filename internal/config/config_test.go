package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		want    *Config
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "valid config",
			file: "testdata/valid.yaml",
			want: &Config{
				Server: Server{
					Address:         "0.0.0.0:9999",
					ShutdownTimeout: 1 * time.Second,
				},
				LogLevel: "debug",
				Storage: struct {
					Postgres      Postgres      `yaml:"postgres"`
					ElasticSearch ElasticSearch `yaml:"elasticsearch"`
					Redis         Redis         `yaml:"redis"`
					RabbitMQ      RabbitMQ      `yaml:"rabbitmq"`
				}{
					Postgres: Postgres{
						Host:     "localhost",
						Port:     5432,
						User:     "user",
						Password: "password",
						DBName:   "test",
					},
					ElasticSearch: ElasticSearch{
						Address: "http://localhost:1234",
					},
					Redis: Redis{
						Address: "localhost:1234",
					},
					RabbitMQ: RabbitMQ{
						Address:       "amqp://user:password@localhost:1234/",
						NoteExchange:  "notes",
						SpaceExchange: "spaces",
					},
				},
				Auth: Auth{
					SecretKey: "a-string-secret-at-least-256-bits-long",
				},
			},
			wantErr: require.NoError,
		},
		{
			name:    "invalid config",
			file:    "testdata/invalid.yaml",
			wantErr: require.Error,
		},
		{
			name:    "non-existent file",
			file:    "testdata/non_existent.yaml",
			wantErr: require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfig(tt.file)
			tt.wantErr(t, err)
			require.Equal(t, tt.want, cfg)
		})
	}
}
