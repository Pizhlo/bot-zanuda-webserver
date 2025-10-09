package main

import (
	"runtime/debug"
	"strings"
	"sync"
)

// BuildInfo - информация о сборке.
type BuildInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
}

// Переменные, инжектируемые при сборке через ldflags
//
//nolint:gochecknoglobals // используются только в этом модуле
var (
	Version   string
	BuildDate string
	GitCommit string

	// Мьютекс для защиты глобальных переменных от data race.
	buildVarsMu sync.RWMutex
)

func ReadBuildInfo() *BuildInfo {
	buildVarsMu.RLock()

	version := Version
	buildDate := BuildDate
	gitCommit := GitCommit

	buildVarsMu.RUnlock()

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return &BuildInfo{
			Name:      "",
			Version:   fallbackVersion(version, buildDate, gitCommit),
			BuildDate: buildDate,
			GitCommit: gitCommit,
		}
	}

	return &BuildInfo{
		Name:      info.Main.Path,
		Version:   fallbackVersion(version, buildDate, gitCommit),
		BuildDate: buildDate,
		GitCommit: gitCommit,
	}
}

func fallbackVersion(version, buildDate, gitCommit string) string {
	// 1) Предпочитаем инжектированную версию
	if strings.TrimSpace(version) != "" {
		return version
	}

	// 2) Формируем версию из Git коммита и даты сборки
	var parts []string

	if gitCommit != "" {
		short := gitCommit
		if len(short) > 7 {
			short = short[:7]
		}

		parts = append(parts, short)
	}

	if buildDate != "" {
		parts = append(parts, buildDate)
	}

	if len(parts) > 0 {
		return strings.Join(parts, "-")
	}

	return "(devel)"
}
