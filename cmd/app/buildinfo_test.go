package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:funlen // длинный тест
func TestReadBuildInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		version         string
		buildDate       string
		gitCommit       string
		expectName      string
		expectVersion   string
		expectBuildDate string
		expectGitCommit string
	}{
		{
			name:            "with all build variables set",
			version:         "v1.2.3",
			buildDate:       "2024-01-15",
			gitCommit:       "abc123def456",
			expectName:      "github.com/example/reviews-backend", // будет заменено на реальный путь
			expectVersion:   "v1.2.3",
			expectBuildDate: "2024-01-15",
			expectGitCommit: "abc123def456",
		},
		{
			name:            "with empty build variables",
			version:         "",
			buildDate:       "",
			gitCommit:       "",
			expectName:      "github.com/example/reviews-backend", // будет заменено на реальный путь
			expectVersion:   "(devel)",
			expectBuildDate: "",
			expectGitCommit: "",
		},
		{
			name:            "with only git commit",
			version:         "",
			buildDate:       "",
			gitCommit:       "abc123def456",
			expectName:      "github.com/example/reviews-backend", // будет заменено на реальный путь
			expectVersion:   "abc123d",                            // первые 7 символов
			expectBuildDate: "",
			expectGitCommit: "abc123def456",
		},
		{
			name:            "with only build date",
			version:         "",
			buildDate:       "2024-01-15",
			gitCommit:       "",
			expectName:      "github.com/example/reviews-backend", // будет заменено на реальный путь
			expectVersion:   "2024-01-15",
			expectBuildDate: "2024-01-15",
			expectGitCommit: "",
		},
		{
			name:            "with git commit and build date",
			version:         "",
			buildDate:       "2024-01-15",
			gitCommit:       "abc123def456",
			expectName:      "github.com/example/reviews-backend", // будет заменено на реальный путь
			expectVersion:   "abc123d-2024-01-15",
			expectBuildDate: "2024-01-15",
			expectGitCommit: "abc123def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Сохраняем оригинальные значения
			originalVersion, originalBuildDate, originalGitCommit := getBuildVars(t)

			// Устанавливаем тестовые значения
			setBuildVars(t, tt.version, tt.buildDate, tt.gitCommit)

			// Восстанавливаем оригинальные значения после теста
			defer setBuildVars(t, originalVersion, originalBuildDate, originalGitCommit)

			result := ReadBuildInfo()

			require.NotNil(t, result, "ReadBuildInfo() returned nil")

			// Проверяем, что Name содержит путь к модулю (может быть разным в разных окружениях)
			require.NotEmpty(t, result.Name, "Expected non-empty Name")
			assert.Equal(t, tt.expectVersion, result.Version)
			assert.Equal(t, tt.expectBuildDate, result.BuildDate)
			assert.Equal(t, tt.expectGitCommit, result.GitCommit)
		})
	}
}

//nolint:funlen // длинный тест
func TestFallbackVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		version   string
		buildDate string
		gitCommit string
		expected  string
	}{
		{
			name:     "with version set",
			version:  "v1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "with version set but whitespace",
			version:  "  v1.2.3  ",
			expected: "  v1.2.3  ", // функция возвращает Version как есть, без обрезки
		},
		{
			name:      "with empty version but git commit",
			version:   "",
			gitCommit: "abc123def456",
			expected:  "abc123d", // первые 7 символов
		},
		{
			name:      "with empty version but build date",
			version:   "",
			buildDate: "2024-01-15",
			expected:  "2024-01-15",
		},
		{
			name:      "with empty version but git commit and build date",
			version:   "",
			gitCommit: "abc123def456",
			buildDate: "2024-01-15",
			expected:  "abc123d-2024-01-15",
		},
		{
			name:      "with short git commit",
			version:   "",
			gitCommit: "abc",
			expected:  "abc",
		},
		{
			name:     "with all empty",
			version:  "",
			expected: "(devel)",
		},
		{
			name:     "with whitespace only",
			version:  "   ",
			expected: "(devel)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Сохраняем оригинальные значения
			originalVersion, originalBuildDate, originalGitCommit := getBuildVars(t)

			// Устанавливаем тестовые значения
			setBuildVars(t, tt.version, tt.buildDate, tt.gitCommit)

			// Восстанавливаем оригинальные значения после теста
			defer setBuildVars(t, originalVersion, originalBuildDate, originalGitCommit)

			result := fallbackVersion(tt.version, tt.buildDate, tt.gitCommit)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadBuildInfoWithDebugInfo(t *testing.T) {
	t.Parallel()

	// Этот тест проверяет, что функция корректно работает с debug.ReadBuildInfo()
	// В реальном окружении debug.ReadBuildInfo() должна возвращать информацию о модуле

	// Сохраняем оригинальные значения
	originalVersion, originalBuildDate, originalGitCommit := getBuildVars(t)

	// Устанавливаем тестовые значения
	setBuildVars(t, "v1.0.0", "2024-01-15", "abc123def456")

	// Восстанавливаем оригинальные значения после теста
	defer setBuildVars(t, originalVersion, originalBuildDate, originalGitCommit)

	result := ReadBuildInfo()

	require.NotNil(t, result, "ReadBuildInfo() returned nil")

	assert.NotEmpty(t, result.Name)
	assert.Equal(t, "v1.0.0", result.Version)
	assert.Equal(t, "2024-01-15", result.BuildDate)
	assert.Equal(t, "abc123def456", result.GitCommit)
}

func TestFallbackVersionEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		version   string
		buildDate string
		gitCommit string
		expected  string
	}{
		{
			name:      "very long git commit",
			version:   "",
			gitCommit: "abcdefghijklmnopqrstuvwxyz1234567890",
			expected:  "abcdefg", // только первые 7 символов
		},
		{
			name:      "git commit with special characters",
			version:   "",
			gitCommit: "abc-123_def",
			expected:  "abc-123",
		},
		{
			name:      "build date with time",
			version:   "",
			buildDate: "2024-01-15T10:30:00Z",
			expected:  "2024-01-15T10:30:00Z",
		},
		{
			name:     "multiple spaces in version",
			version:  "   v1.2.3   ",
			expected: "   v1.2.3   ", // функция возвращает Version как есть, без обрезки
		},
		{
			name:     "tabs in version",
			version:  "\tv1.2.3\t",
			expected: "\tv1.2.3\t", // функция возвращает Version как есть, без обрезки
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Сохраняем оригинальные значения
			originalVersion, originalBuildDate, originalGitCommit := getBuildVars(t)

			// Устанавливаем тестовые значения
			setBuildVars(t, tt.version, tt.buildDate, tt.gitCommit)

			// Восстанавливаем оригинальные значения после теста
			defer setBuildVars(t, originalVersion, originalBuildDate, originalGitCommit)

			result := fallbackVersion(tt.version, tt.buildDate, tt.gitCommit)

			assert.Equal(t, tt.expected, result)
		})
	}
}

// setBuildVars устанавливает значения глобальных переменных сборки (только для тестов).
func setBuildVars(t *testing.T, version, buildDate, gitCommit string) {
	t.Helper()

	buildVarsMu.Lock()
	defer buildVarsMu.Unlock()

	Version = version
	BuildDate = buildDate
	GitCommit = gitCommit
}

// getBuildVars возвращает текущие значения глобальных переменных сборки (только для тестов).
func getBuildVars(t *testing.T) (string, string, string) {
	t.Helper()

	buildVarsMu.RLock()
	defer buildVarsMu.RUnlock()

	return Version, BuildDate, GitCommit
}
