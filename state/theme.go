package state

import (
	"os"
	"runtime"
	"strings"
)

// ProgressTheme contains all the characters we need to show progress
type ProgressTheme struct {
	BarStart        string
	BarEnd          string
	Current         string
	CurrentHalfTone string
	Empty           string
	OpSign          string
	StatSign        string
	Separator       string
}

var themes = map[string]*ProgressTheme{
	"unicode": {"▐", "▌", "▓", "▒", "░", "•", "✓", "•"},
	"ascii":   {"|", "|", "#", "=", "-", ">", "<", "|"},
	"cp437":   {"▐", "▌", "█", "▒", "░", "∙", "√", "∙"},
}

// EnableBeepsForAdam is there for backwards compatibility, but mostly, fun
func EnableBeepsForAdam() {
	// this character emits a system bell sound. Adam loves it.
	themes["cp437"].OpSign = "•"
}

func getCharset() string {
	if runtime.GOOS == "windows" && os.Getenv("OS") != "CYGWIN" {
		return "cp437"
	}

	for _, envkey := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		envval := os.Getenv(envkey)
		if strings.HasSuffix(envval, ".UTF-8") || strings.HasSuffix(envval, ".utf8") {
			return "unicode"
		}
	}

	return "ascii"
}

var theme = themes[getCharset()]

// GetTheme returns the theme used to show progress
func GetTheme() *ProgressTheme {
	return theme
}
