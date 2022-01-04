package util

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// EnsureDir checks for existence of a directory and recursively creates it if needed
func EnsureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}
}

// GetAppDir forms the application data directory from the given input
func GetAppDir(baseDir string, appName string) string {
	return path.Join(baseDir, appName)
}

// GetHomeDir gets the user's home directory with special casing for windows
func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("There was a problem finding the user's home directory")
	}

	if runtime.GOOS == "windows" {
		home = path.Join(home, "AppData")
	}

	return home
}

// InitBaseDir creates the base directory
func InitBaseDir(baseDir string) string {
	if !filepath.IsAbs(baseDir) {
		homedir := GetHomeDir()
		baseDir = filepath.Join(homedir, baseDir)
	}
	EnsureDir(baseDir)

	return baseDir
}
