package util

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// EnsureDir checks for existence of a directory and recursively creates it if needed
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	}

	return nil
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
func InitBaseDir(baseDir string) (string, error) {
	if !filepath.IsAbs(baseDir) {
		homedir := GetHomeDir()
		baseDir = filepath.Join(homedir, baseDir)
	}
	if err := EnsureDir(baseDir); err != nil {
		return "", err
	}

	return baseDir, nil
}
